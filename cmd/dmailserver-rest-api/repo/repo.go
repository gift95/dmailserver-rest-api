package repo

import (
	"errors"
	"os/exec"
	"strings"
	"regexp"
	"log" 
	"github.com/pjotrscholtze/dmailserver-rest-api/models"
)

type setupRepo struct {
	commandPrefix string
}
type SetupRepo interface {
	ListEmail() ([]*models.EmailAccountListItem, error)
	AddEmail(account models.EmailAccount) error
	UpdateEmail(account models.EmailAccount) error
	RemoveEmail(emailAddress string) error
	HasEmail(emailAddress string) (bool, error)

	GetEmail(addres string) (*models.EmailAccountListItem, error)
	HasAlias(aliasAddress, recepientAddres string) (bool, error)
	AddAlias(aliasAddress, recepientAddres string) error
	RemoveAlias(aliasAddress, recepientAddres string) error

	AddressExistsOnServer(addres string) (bool, error)
	ListFail2ban() (models.Fail2banListItem, error)
	RemoveFail2ban(ip string) error
	AddFail2ban(ip string) error
	HasFail2banIp(ip string) (bool, error)
}

func (sr *setupRepo) AddressExistsOnServer(addres string) (bool, error) {
	emails, err := sr.ListEmail()
	if err != nil {
		return false, err
	}
	for _, mainEmail := range emails {
		if mainEmail.Address == addres {
			return true, nil
		}
		for _, alias := range mainEmail.Aliases {
			if alias == addres {
				return true, nil
			}
		}
	}
	return false, nil
}

func (sr *setupRepo) AddAlias(aliasAddress, recepientAddres string) error {
	if aliasAddress == "" {
		return errors.New("A alias is required for an email account")
	}

	if recepientAddres == "" {
		return errors.New("A recepient is required for an email account")
	}

	recepientExists, err := sr.AddressExistsOnServer(recepientAddres)
	if err != nil {
		return err
	}
	if !recepientExists {
		return errors.New("Recepient address does not exist on this server!")
	}
	aliasExists, err := sr.AddressExistsOnServer(aliasAddress)
	if err != nil {
		return err
	}
	if aliasExists {
		return errors.New("Alias already in use!")
	}

	parts := strings.Split(sr.commandPrefix+"setup alias add", " ")
	parts = append(parts, aliasAddress)
	parts = append(parts, recepientAddres)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err = proc.Output()
	return err
}
func (sr *setupRepo) GetEmail(addres string) (*models.EmailAccountListItem, error) {
	emails, err := sr.ListEmail()
	if err != nil {
		return nil, err
	}
	for i := range emails {
		if emails[i].Address == addres {
			return emails[i], nil
		}
	}
	return nil, errors.New("Email account not found!")
}
func (sr *setupRepo) HasAlias(aliasAddress, recepientAddres string) (bool, error) {
	emailAcc, err := sr.GetEmail(recepientAddres)
	if err != nil {
		return false, err
	}
	for _, address := range emailAcc.Aliases {
		if address == aliasAddress {
			return true, nil
		}
	}
	return false, nil
}

func (sr *setupRepo) RemoveAlias(aliasAddress, recepientAddres string) error {
	if aliasAddress == "" {
		return errors.New("A alias is required for an email account")
	}

	if recepientAddres == "" {
		return errors.New("A recepient is required for an email account")
	}

	recepientExists, err := sr.AddressExistsOnServer(recepientAddres)
	if err != nil {
		return err
	}
	if !recepientExists {
		return errors.New("Recepient address does not exist on this server!")
	}
	aliasExists, err := sr.AddressExistsOnServer(aliasAddress)
	if err != nil {
		return err
	}
	if !aliasExists {
		return errors.New("Alias does not exist!")
	}

	parts := strings.Split(sr.commandPrefix+"setup alias del", " ")
	parts = append(parts, aliasAddress)
	parts = append(parts, recepientAddres)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err = proc.Output()
	return err
}

func (sr *setupRepo) ListEmail() ([]*models.EmailAccountListItem, error) {
	parts := strings.Split(sr.commandPrefix+"setup email list", " ")
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	t, err := proc.Output()
	if err != nil {
		return nil, err
	}

	emailAddresses := []*models.EmailAccountListItem{}
	for _, line := range strings.Split(string(t), "\n") {
		if len(line) < 3 {
			continue
		}

		if line[0] == '*' {
			// 解析主邮箱行
			lineParts := strings.Split(line[2:], " ( ")
			if len(lineParts) < 2 {
				continue // 格式异常，跳过
			}
			
			lineSubParts := strings.Split(lineParts[1], " / ")
			if len(lineSubParts) < 2 {
				continue // 格式异常，跳过
			}
			
			usage := lineSubParts[0]
			lineSubParts2 := strings.Split(lineSubParts[1], " ) [")
			if len(lineSubParts2) < 2 {
				continue // 格式异常，跳过
			}
			
			// 安全获取使用百分比
			usagePercentage := ""
			if len(lineSubParts2[1]) > 1 {
				usagePercentage = lineSubParts2[1][:len(lineSubParts2[1])-1]
			}
			
			emailAddresses = append(emailAddresses, &models.EmailAccountListItem{
				Address: lineParts[0],
				Aliases: []string{},
				Quota: &models.Quota{
					Usage:           usage,
					Limit:           lineSubParts2[0],
					UsagePercentage: usagePercentage,
				},
			})
		} else {
			// 修复点：解析别名行时添加安全检查
			if !strings.Contains(line, "[ aliases -> ") {
				continue // 不是别名行，跳过
			}
			
			aliasParts := strings.Split(line, "[ aliases -> ")
			if len(aliasParts) < 2 {
				continue // 别名格式异常，跳过
			}
			
			reducedLine := aliasParts[1]
			if len(reducedLine) < 2 {
				continue // 别名部分太短，跳过
			}
			
			reducedLine = reducedLine[:len(reducedLine)-2]
			
			// 确保有邮箱可以添加别名
			if len(emailAddresses) == 0 {
				continue // 没有主邮箱，跳过
			}
			
			// 分割别名并过滤空值
			aliases := strings.Split(reducedLine, ", ")
			validAliases := []string{}
			for _, alias := range aliases {
				if strings.TrimSpace(alias) != "" {
					validAliases = append(validAliases, alias)
				}
			}
			
			emailAddresses[len(emailAddresses)-1].Aliases = validAliases
		}
	}

	return emailAddresses, nil
}
func (sr *setupRepo) AddEmail(account models.EmailAccount) error {
	if account.Password == "" {
		return errors.New("A password is required for an email account")
	}

	parts := strings.Split(sr.commandPrefix+"setup email add", " ")
	parts = append(parts, account.Address)
	parts = append(parts, account.Password)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err := proc.Output()
	if err != nil {
		return err
	}

	return nil
}
func (sr *setupRepo) UpdateEmail(account models.EmailAccount) error {
	if account.Password == "" {
		return errors.New("A password is required for an email account")
	}

	parts := strings.Split(sr.commandPrefix+"setup email update", " ")
	parts = append(parts, account.Address)
	parts = append(parts, account.Password)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err := proc.Output()
	if err != nil {
		return err
	}

	return nil
}
func (sr *setupRepo) RemoveEmail(emailAddress string) error {
	parts := strings.Split(sr.commandPrefix+"setup email del", " ")
	parts = append(parts, emailAddress)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err := proc.Output()
	if err != nil {
		return err
	}

	return nil
}
// ListFail2ban 获取 Fail2ban 封禁列表
func (sr *setupRepo) ListFail2ban() (models.Fail2banListItem, error) {
	// 初始化返回结构
	out := models.Fail2banListItem{
		BannedInPostfix:     []string{},
		BannedInDovecot:     []string{},
		BannedInCustom: []string{},
	}
	
	// 构建命令
	cmdStr := strings.TrimSpace(sr.commandPrefix + "setup fail2ban")
	if cmdStr == "" {
		return out, errors.New("command prefix is empty")
	}
	
	// 安全地分割命令
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return out, errors.New("invalid command")
	}
	
	log.Printf("[DEBUG] Executing fail2ban command: %v", parts)
	
	// 执行命令
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[DEBUG] Fail2ban command completed with error (may be normal): %v, output: %s", err, string(output))
		return out, nil
	}
	
	log.Printf("[DEBUG] Fail2ban raw output: %s", string(output))
	
	// 解析输出
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		log.Printf("[DEBUG] Processing line: %s", line)
		
		// 检查是否是 Custom 封禁（新增）
		if strings.Contains(strings.ToLower(line), "custom") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ipPart := strings.TrimSpace(parts[1])
				ips := strings.Fields(ipPart)
				for _, ip := range ips {
					if isValidIP(ip) {
						out.BannedInCustom = append(out.BannedInCustom, ip)
					}
				}
				log.Printf("[DEBUG] Found Custom bans: %v", out.BannedInCustom)
			}
		}
		
		// 检查是否是 Dovecot 封禁
		if strings.Contains(strings.ToLower(line), "dovecot") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ipPart := strings.TrimSpace(parts[1])
				ips := strings.Fields(ipPart)
				for _, ip := range ips {
					if isValidIP(ip) {
						out.BannedInDovecot = append(out.BannedInDovecot, ip)
					}
				}
				log.Printf("[DEBUG] Found Dovecot bans: %v", out.BannedInDovecot)
			}
		}
		
		// 检查是否是 Postfix 封禁
		if strings.Contains(strings.ToLower(line), "postfix") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ipPart := strings.TrimSpace(parts[1])
				ips := strings.Fields(ipPart)
				for _, ip := range ips {
					if isValidIP(ip) {
						out.BannedInPostfix = append(out.BannedInPostfix, ip)
					}
				}
				log.Printf("[DEBUG] Found Postfix bans: %v", out.BannedInPostfix)
			}
		}
		
	}
	
	// 去重
	out.BannedInPostfix = uniqueStrings(out.BannedInPostfix)
	out.BannedInDovecot = uniqueStrings(out.BannedInDovecot)
	out.BannedInCustom = uniqueStrings(out.BannedInCustom)
	
	log.Printf("[DEBUG] Final fail2ban list: Postfix=%v, Dovecot=%v, SASL=%v, Custom=%v", 
		out.BannedInPostfix, out.BannedInDovecot, out.BannedInCustom)
	
	return out, nil
}

// uniqueStrings 去重字符串切片
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, str := range input {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	
	return result
}

// isValidIP 简单验证 IP 地址格式
func isValidIP(ip string) bool {
	// 使用正则表达式验证 IP 格式
	ipPattern := `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`
	matched, _ := regexp.MatchString(ipPattern, ip)
	return matched
}

func (sr *setupRepo) AddFail2ban(ip string) error {
	parts := strings.Split(sr.commandPrefix+"setup fail2ban ban", " ")
	parts = append(parts, ip)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err := proc.Output()
	if err != nil {
		return err
	}

	return nil
}
func (sr *setupRepo) RemoveFail2ban(ip string) error {
	parts := strings.Split(sr.commandPrefix+"setup fail2ban unban", " ")
	parts = append(parts, ip)
	proc := exec.Command(parts[0], parts[1:]...)
	proc.Wait()
	_, err := proc.Output()
	if err != nil {
		return err
	}

	return nil
}
func (sr *setupRepo) HasEmail(emailAddress string) (bool, error) {
	emails, err := sr.ListEmail()
	if err != nil {
		return false, err
	}

	for _, email := range emails {
		if email.Address == emailAddress {
			return true, nil
		}
	}

	return false, nil
}
func (sr *setupRepo) HasFail2banIp(ip string) (bool, error) {
	bans, err := sr.ListFail2ban()
	if err != nil {
		return false, err
	}

	for _, currentIp := range bans.BannedInDovecot {
		if currentIp == ip {
			return true, nil
		}
	}

	for _, currentIp := range bans.BannedInPostfix {
		if currentIp == ip {
			return true, nil
		}
	}

	for _, currentIp := range bans.BannedInCustom {
		if currentIp == ip {
			return true, nil
		}
	}

	return false, nil
}

func NewSetupRepo(commandPrefix string) SetupRepo {
	return &setupRepo{
		commandPrefix: commandPrefix,
	}
}
