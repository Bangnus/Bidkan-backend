package email_auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"html"
	"math/big"
	"net/url"
	"os"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/google/uuid"
)

func generateSixDigitOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func resolveTenantID(ctx context.Context, tenantRepo repository.TenantRepository, email, campusSlug string) *uuid.UUID {
	if tenantRepo == nil {
		return nil
	}
	lowerEmail := strings.ToLower(strings.TrimSpace(email))
	slug := strings.ToLower(strings.TrimSpace(campusSlug))
	switch {
	case strings.HasSuffix(lowerEmail, "ku.th") || strings.HasSuffix(lowerEmail, "ku.ac.th"):
		slug = "ku"
	case strings.HasSuffix(lowerEmail, "cmu.ac.th"):
		slug = "cmu"
	case slug == "":
		slug = "ku"
	}
	if t, err := tenantRepo.GetBySlug(ctx, slug); err == nil && t != nil {
		id := t.ID
		return &id
	}
	if fallback, err := tenantRepo.GetBySlug(ctx, "ku"); err == nil && fallback != nil {
		id := fallback.ID
		return &id
	}
	return nil
}

func buildVerificationEmail(toEmail, username, otp string) domainService.EmailMessage {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_APP_URL")), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_CLIENT_URL")), "/")
	}
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	verifyURL := fmt.Sprintf("%s/login?verify_email=%s&code=%s", baseURL, url.QueryEscape(toEmail), url.QueryEscape(otp))
	safeName := html.EscapeString(username)
	safeOTP := html.EscapeString(otp)
	safeURL := html.EscapeString(verifyURL)

	text := fmt.Sprintf(
		"สวัสดีคุณ %s,\n\nรหัสยืนยันอีเมล BidKan Campus ของคุณคือ: %s (มีอายุ 15 นาที)\n\nหรือคลิกลิงก์นี้เพื่อยืนยันทันที:\n%s\n\nหากคุณไม่ได้เป็นผู้สมัครสมาชิก สามารถละเว้นอีเมลฉบับนี้ได้",
		username, otp, verifyURL,
	)

	htmlBody := fmt.Sprintf(`<!doctype html>
<html lang="th"><body style="margin:0;padding:24px;background:#F4F5F9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#1E293B;">
<div style="max-width:480px;margin:0 auto;background:#FFFFFF;border-radius:20px;padding:32px;border:1px solid #E9D5FF;">
  <div style="font-size:13px;font-weight:700;color:#7C3AED;margin-bottom:8px;">BidKan Campus</div>
  <h1 style="margin:0 0 12px;font-size:20px;color:#0F172A;">ยืนยันอีเมลของคุณ</h1>
  <p style="margin:0 0 20px;font-size:14px;line-height:1.6;color:#475569;">สวัสดีคุณ <strong>%s</strong> กรุณานำรหัส OTP ด้านล่างไปกรอกเพื่อเปิดใช้งานบัญชี BidKan (รหัสมีอายุ 15 นาที)</p>
  <div style="text-align:center;padding:18px;background:#F5F3FF;border:1px dashed #8B5CF6;border-radius:14px;margin-bottom:22px;">
    <span style="font-size:32px;font-weight:800;letter-spacing:8px;color:#6D28D9;">%s</span>
  </div>
  <div style="text-align:center;margin-bottom:20px;">
    <a href="%s" style="display:inline-block;background:#7C3AED;color:#FFFFFF;text-decoration:none;font-weight:700;font-size:14px;padding:12px 24px;border-radius:12px;">คลิกเพื่อยืนยันอีเมลอัตโนมัติ</a>
  </div>
  <p style="margin:0;font-size:12px;color:#94A3B8;">หากคุณไม่ได้ทำการสมัครสมาชิก กรุณาเพิกเฉยต่ออีเมลฉบับนี้</p>
</div></body></html>`, safeName, safeOTP, safeURL)

	return domainService.EmailMessage{
		To:      toEmail,
		Subject: "รหัสยืนยันอีเมล BidKan Campus (" + otp + ")",
		Text:    text,
		HTML:    htmlBody,
	}
}
