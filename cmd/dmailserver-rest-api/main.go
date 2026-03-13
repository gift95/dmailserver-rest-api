package main

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"html/template"
	"net/http"
	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	"github.com/pjotrscholtze/dmailserver-rest-api/cmd/dmailserver-rest-api/cnf"
	"github.com/pjotrscholtze/dmailserver-rest-api/cmd/dmailserver-rest-api/controller"
	"github.com/pjotrscholtze/dmailserver-rest-api/cmd/dmailserver-rest-api/repo"
	"github.com/pjotrscholtze/dmailserver-rest-api/models"
	"github.com/pjotrscholtze/dmailserver-rest-api/restapi"
	"github.com/pjotrscholtze/dmailserver-rest-api/restapi/operations"
)

func main() {
	configPath := "../../config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := cnf.GetConfig(configPath)
	if err != nil {
		panic(err)
	}
	_ = config

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("Starting DMailserver-rest-api")

	slog.Info("Setting command prefix", "commandPrefix", config.ServerConfig.CommandPrefix)
	sr := repo.NewSetupRepo(config.ServerConfig.CommandPrefix)

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewDmailserverRestAPIAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Swagger Petstore - OpenAPI 3.0"
	parser.LongDescription = swaggerSpec.Spec().Info.Description
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	// Applies when the "x-token" header is set
	api.APIKeyAuth = func(token string) (interface{}, error) {
		if token == config.ServerConfig.APIKey { //"abcdefuvwxyz" {
			prin := models.Principal(token)
			return &prin, nil
		}
		api.Logger("Access attempt with incorrect api key auth: %s", token)
		return nil, errors.New("incorrect api key auth")
	}
	controller.SetupController(api, sr)
	server.Port = config.ServerConfig.Port
	server.Host = config.ServerConfig.Host
	slog.Info("Setting up port and host for connection listening", "port", server.Port, "host", server.Host)
	server.ConfigureAPI()
	// 使用中间件包装，先检查静态文件，再交给API
	originalHandler := server.GetHandler()
	server.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试提供静态文件
		if r.URL.Path != "/api" && !strings.HasPrefix(r.URL.Path, "/api/") {
			// 检查是否是HTML文件
			if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" {
				// 构建文件路径
				filePath := "./static"
				if r.URL.Path != "/" {
					filePath += r.URL.Path
				} else {
					filePath += "/index.html"
				}
				
				// 检查文件是否存在
				if _, err := os.Stat(filePath); err == nil {
					// 准备模板数据
						data := map[string]interface{}{
							"MAIL_HOST": config.ServerConfig.MailHost,
						}
					
					// 解析并渲染模板
					tmpl, err := template.ParseFiles(filePath)
					if err != nil {
						// 渲染失败，直接提供文件
						staticHandler := http.FileServer(http.Dir("./static"))
						staticHandler.ServeHTTP(w, r)
						return
					}
					err = tmpl.Execute(w, data)
					if err != nil {
						// 渲染失败，直接提供文件
						staticHandler := http.FileServer(http.Dir("./static"))
						staticHandler.ServeHTTP(w, r)
						return
					}
					return
				}
			}
			// 其他静态文件直接提供
			staticHandler := http.FileServer(http.Dir("./static"))
			staticHandler.ServeHTTP(w, r)
			return
		}
		// 其他请求交给API处理
		originalHandler.ServeHTTP(w, r)
	}))


	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
