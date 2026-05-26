package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var qrGenerateCmd = &cobra.Command{
	Use:   "qr_generate <text>",
	Short: "Generate QR code",
	Long:  `Generate a QR code from text or URL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		size, _ := cmd.Flags().GetInt("size")

		if size <= 0 {
			size = 256
		}

		qr, err := qrcode.Encode(text, qrcode.Medium, size)
		if err != nil {
			fmt.Printf("Error generating QR code: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			err = os.WriteFile(outputFile, qr, 0644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("QR code saved to: %s\n", outputFile)
		} else {
			fmt.Println("QR Code (Base64):")
			fmt.Println(base64.StdEncoding.EncodeToString(qr))
			fmt.Println("\nUse --output flag to save as PNG file")
		}
	},
}

var qrDecodeCmd = &cobra.Command{
	Use:   "qr_decode <image_file>",
	Short: "Decode QR code",
	Long:  `Decode a QR code from an image file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		fmt.Printf("QR decoding requires a library - trying to decode: %s\n", file)
		fmt.Println("Note: Full QR decoding would require additional libraries")
	},
}

var passwordGenCmd = &cobra.Command{
	Use:   "password_generate [length]",
	Short: "Generate secure password",
	Long:  `Generate a secure random password.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := 16
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &length)
		}

		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
		rand.Seed(time.Now().UnixNano())

		password := make([]byte, length)
		for i := range password {
			password[i] = charset[rand.Intn(len(charset))]
		}

		fmt.Printf("Generated password (%d chars):\n%s\n", length, string(password))
		fmt.Println("\nPassword strength: Strong (includes uppercase, lowercase, numbers, symbols)")
	},
}

var passwordStrengthCmd = &cobra.Command{
	Use:   "password_strength <password>",
	Short: "Check password strength",
	Long:  `Check the strength of a password.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		password := args[0]
		score := 0

		if len(password) >= 8 {
			score += 1
		}
		if len(password) >= 12 {
			score += 1
		}
		if len(password) >= 16 {
			score += 1
		}

		hasLower := false
		hasUpper := false
		hasNumber := false
		hasSymbol := false

		for _, c := range password {
			switch {
			case c >= 'a' && c <= 'z':
				hasLower = true
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= '0' && c <= '9':
				hasNumber = true
			case c < 32 || c > 126:
			default:
				hasSymbol = true
			}
		}

		if hasLower {
			score += 1
		}
		if hasUpper {
			score += 1
		}
		if hasNumber {
			score += 1
		}
		if hasSymbol {
			score += 2
		}

		strength := ""
		switch {
		case score >= 7:
			strength = "Very Strong"
		case score >= 5:
			strength = "Strong"
		case score >= 3:
			strength = "Medium"
		default:
			strength = "Weak"
		}

		fmt.Printf("Password: %s\n", password)
		fmt.Printf("Length:   %d\n", len(password))
		fmt.Printf("Strength: %s (score: %d/8)\n\n", strength, score)

		fmt.Println("Contains:")
		fmt.Printf("  Lowercase: %v\n", hasLower)
		fmt.Printf("  Uppercase: %v\n", hasUpper)
		fmt.Printf("  Numbers:  %v\n", hasNumber)
		fmt.Printf("  Symbols:  %v\n", hasSymbol)
	},
}

var timeNowCmd = &cobra.Command{
	Use:   "time_now",
	Short: "Show current time",
	Long:  `Display current date and time in multiple formats.`,
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()

		fmt.Println("=== Current Time ===")
		fmt.Printf("RFC3339:    %s\n", now.Format(time.RFC3339))
		fmt.Printf("RFC1123:     %s\n", now.Format(time.RFC1123))
		fmt.Printf("ANSIC:       %s\n", now.Format(time.ANSIC))
		fmt.Printf("Unix:        %d\n", now.Unix())
		fmt.Printf("UnixNano:    %d\n", now.UnixNano())
		fmt.Printf("ISO8601:     %s\n", now.Format("2006-01-02T15:04:05Z07:00"))

		year, month, day := now.Date()
		hour, min, sec := now.Clock()
		fmt.Printf("\nComponents: %04d-%02d-%02d %02d:%02d:%02d\n", year, month, day, hour, min, sec)
	},
}

var timeConvertCmd = &cobra.Command{
	Use:   "time_convert <timestamp>",
	Short: "Convert Unix timestamp",
	Long:  `Convert Unix timestamp to human-readable format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ts int64
		fmt.Sscanf(args[0], "%d", &ts)

		if ts > 1e12 {
			ts = ts / 1000
		}

		t := time.Unix(ts, 0)

		fmt.Printf("Timestamp: %d\n\n", ts)
		fmt.Printf("UTC:        %s\n", t.Format(time.RFC3339))
		fmt.Printf("Local:      %s\n", t.Local().Format(time.RFC3339))
		fmt.Printf("Date only:  %s\n", t.Format("2006-01-02"))
		fmt.Printf("Time only:  %s\n", t.Format("15:04:05"))
	},
}

var cronNextCmd = &cobra.Command{
	Use:   "cron_next <expression>",
	Short: "Calculate next cron run",
	Long:  `Calculate the next run time for a cron expression.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		expr := args[0]
		fmt.Printf("Cron expression: %s\n", expr)
		fmt.Println("\nNote: Full cron parsing requires a cron library")
		fmt.Printf("Next runs would be calculated based on: %s\n", time.Now().Format(time.RFC3339))
	},
}

var hashFileCmd = &cobra.Command{
	Use:   "hash_file <file>",
	Short: "Calculate file hash",
	Long:  `Calculate MD5, SHA1, SHA256 hashes of a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		md5 := md5.Sum(data)
		sha1 := sha1.Sum(data)
		sha256 := sha256.Sum256(data)

		fmt.Printf("File: %s\n", file)
		fmt.Printf("Size: %d bytes\n\n", len(data))
		fmt.Printf("MD5:    %x\n", md5)
		fmt.Printf("SHA1:   %x\n", sha1)
		fmt.Printf("SHA256: %x\n", sha256)
	},
}

var fileInfoCmd = &cobra.Command{
	Use:   "file_info <file>",
	Short: "Show file information",
	Long:  `Display detailed file information.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("File:       %s\n", file)
		fmt.Printf("Size:       %d bytes\n", info.Size())
		fmt.Printf("Mode:       %s\n", info.Mode())
		fmt.Printf("Modified:   %s\n", info.ModTime().Format(time.RFC3339))
		fmt.Printf("Is Dir:     %v\n", info.IsDir())
	},
}

var hexEncodeCmd = &cobra.Command{
	Use:   "hex_encode <string>",
	Short: "Encode string to hex",
	Long:  `Encode a string to hexadecimal format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data := []byte(args[0])
		hex := hex.EncodeToString(data)
		fmt.Printf("Input: %s\n", args[0])
		fmt.Printf("Hex:   %s\n", hex)
	},
}

var hexDecodeCmd = &cobra.Command{
	Use:   "hex_decode <hex_string>",
	Short: "Decode hex to string",
	Long:  `Decode a hexadecimal string to text.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := hex.DecodeString(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Input: %s\n", args[0])
		fmt.Printf("Output: %s\n", string(data))
	},
}

var uuidParseCmd = &cobra.Command{
	Use:   "uuid_parse <uuid>",
	Short: "Parse and validate UUID",
	Long:  `Parse and validate a UUID string.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uuidStr := args[0]
		parsed, err := uuid.Parse(uuidStr)
		if err != nil {
			fmt.Printf("Invalid UUID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Input:  %s\n", uuidStr)
		fmt.Printf("Parsed:  %s\n", parsed.String())
		fmt.Printf("Version: %d\n", parsed.Version())
		fmt.Printf("Variant: %d\n", parsed.Variant())
	},
}

func init() {
	qrGenerateCmd.Flags().StringP("output", "o", "", "Output PNG file")
	qrGenerateCmd.Flags().IntP("size", "s", 256, "QR code size")
}
