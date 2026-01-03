package pkg

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailSender 邮件发送器
type EmailSender struct {
	config *EmailConfig
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(config *EmailConfig) *EmailSender {
	return &EmailSender{
		config: config,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *EmailSender) SendVerificationCode(toEmail, code string) error {
	subject := "邮箱验证码"
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px; }
				.container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
				.header { text-align: center; color: #333; margin-bottom: 30px; }
				.code-box { background-color: #f8f9fa; border: 2px dashed #007bff; border-radius: 5px; padding: 20px; text-align: center; margin: 20px 0; }
				.code { font-size: 32px; font-weight: bold; color: #007bff; letter-spacing: 5px; }
				.info { color: #666; font-size: 14px; line-height: 1.6; margin-top: 20px; }
				.footer { text-align: center; color: #999; font-size: 12px; margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>邮箱验证码</h2>
				</div>
				<p>您好，</p>
				<p>您正在进行邮箱验证操作，您的验证码是：</p>
				<div class="code-box">
					<div class="code">%s</div>
				</div>
				<div class="info">
					<p>• 验证码有效期为 <strong>5分钟</strong>，请尽快使用</p>
					<p>• 如果这不是您本人的操作，请忽略此邮件</p>
					<p>• 请勿将验证码透露给他人</p>
				</div>
				<div class="footer">
					<p>此邮件由系统自动发送，请勿回复</p>
					<p>© Smart Collab Gallery</p>
				</div>
			</div>
		</body>
		</html>
	`, code)

	return s.sendEmail(toEmail, subject, htmlBody)
}

// sendEmail 发送邮件
func (s *EmailSender) sendEmail(to, subject, htmlBody string) error {
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", m.FormatAddress(s.config.FromEmail, s.config.FromName))

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置邮件正文（HTML格式）
	m.SetBody("text/html", htmlBody)

	// 创建 SMTP 拨号器
	d := gomail.NewDialer(
		s.config.SMTPHost,
		s.config.SMTPPort,
		s.config.SMTPUser,
		s.config.SMTPPassword,
	)

	// 对于某些 SMTP 服务器，需要设置 TLS 配置
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.config.SMTPHost,
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}
