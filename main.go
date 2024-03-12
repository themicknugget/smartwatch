package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Load environment variables from file if ENVFILE is specified and the file exists
	envFilePath := os.Getenv("ENVFILE")
	if envFilePath != "" {
		if _, err := os.Stat(envFilePath); err == nil {
			if err := loadEnvFromFile(envFilePath); err != nil {
				log.Fatalf("Failed to load environment variables from file: %s", err)
			}
		} else {
			log.Printf("Specified ENVFILE does not exist or cannot be accessed: %s", envFilePath)
		}
	}

	// Validate required environment variables
	requiredVars := []string{"SMTP_SERVER", "SMTP_PORT", "SENDER_EMAIL", "SENDER_PASSWORD", "RECIPIENT_EMAIL", "DEVICES"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Required environment variable %s is not set. Please check your configuration.", v)
		}
	}

	// Configuration variables
	checkInterval := getEnvAsDuration("CHECK_INTERVAL", 1*time.Hour)
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	senderEmail := os.Getenv("SENDER_EMAIL")
	senderPassword := os.Getenv("SENDER_PASSWORD")
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")
	smartctlLocation := os.Getenv("SMARTCTL_LOCATION")
	devices := getEnvAsSlice("DEVICES", ",", []string{"/dev/sda"}) // Default to /dev/sda if not specified

	for _, device := range devices {
		checkAndSendEmail(smtpServer, smtpPort, senderEmail, senderPassword, recipientEmail, smartctlLocation, device)
		time.Sleep(checkInterval)
	}
}

func loadEnvFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") { // Skip empty lines and comments
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Ignoring invalid line in env file: %s", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func checkAndSendEmail(smtpServer, smtpPort, senderEmail, senderPassword, recipientEmail, smartctlLocation, device string) {
	var cmd *exec.Cmd
	if strings.HasPrefix(device, "/dev/nvme") {
		// Use the '-x' option for NVMe devices for extended information (as an example)
		cmd = exec.Command(smartctlLocation, "-x", device)
	} else {
		// Use '-H' option for traditional SATA devices
		cmd = exec.Command(smartctlLocation, "-H", device)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running smartctl on %s: %v", device, err)
		return
	}

	if strings.Contains(out.String(), "PASSED") || strings.Contains(out.String(), "SMART Health Status: OK") {
		fmt.Printf("Disk %s is healthy. No action required.\n", device)
	} else {
		fmt.Printf("Warning/Error found on %s. Sending email...\n", device)
		sendEmail(smtpServer, smtpPort, senderEmail, senderPassword, recipientEmail, device+":\n"+out.String())
	}
}

func sendEmail(smtpServer, smtpPort, senderEmail, senderPassword, recipientEmail, body string) {
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)
	to := []string{recipientEmail}
	msg := []byte("To: " + recipientEmail + "\r\n" +
		"Subject: SMART Disk Warning/Error on " + strings.Split(body, ":\n")[0] + "\r\n" +
		"\r\n" +
		body)

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, senderEmail, to, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		fmt.Println("Email sent successfully.")
	}
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Invalid format for duration in environment variable %s: %v", key, err)
		return defaultVal
	}
	return duration
}

func getEnvAsSlice(key string, delimiter string, defaultVal []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return strings.Split(value, delimiter)
}
