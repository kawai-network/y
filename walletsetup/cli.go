package walletsetup

import (
	"fmt"
	"os"
	"strings"
)

// UI is a minimal interface for CLI output.
type UI interface {
	Println(a ...any)
	Printf(format string, a ...any)
}

// WalletOps defines the wallet operations required by the setup flow.
type WalletOps interface {
	HasWallet() bool
	GenerateMnemonic() (string, error)
	CreateWallet(password string, mnemonic string, description string) (string, error)
	ImportKeystore(keystoreJSON string, password string, description string) (string, error)
	ImportPrivateKey(privateKeyHex string, password string, description string) (string, error)
}

// WalletResult represents wallet setup result.
type WalletResult struct {
	Address string
	Created bool
}

// SetupOptions controls optional behavior.
type SetupOptions struct {
	WalletName       string
	CopyMnemonic     func(string) error
	ValidateMnemonic func(string) (bool, string)
	ValidateKeystore func(string) (bool, string)
	ValidatePrivKey  func(string) (bool, string)
	PasswordStrength func(string) (int, string, string)
}

// RunInteractiveSetup runs the interactive wallet setup flow.
func RunInteractiveSetup(ui UI, input *InputReader, wallet WalletOps, opts SetupOptions) (*WalletResult, error) {
	if ui == nil {
		return nil, fmt.Errorf("ui is required")
	}
	if input == nil {
		return nil, fmt.Errorf("input reader is required")
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet ops is required")
	}

	if opts.WalletName == "" {
		opts.WalletName = "My Wallet"
	}
	if opts.ValidateMnemonic == nil {
		opts.ValidateMnemonic = ValidateMnemonicBasic
	}
	if opts.ValidateKeystore == nil {
		opts.ValidateKeystore = ValidateKeystoreJSON
	}
	if opts.ValidatePrivKey == nil {
		opts.ValidatePrivKey = ValidatePrivateKey
	}
	if opts.PasswordStrength == nil {
		opts.PasswordStrength = PasswordStrength
	}

	if wallet.HasWallet() {
		ui.Println("🔐 Existing Wallet Detected")
		ui.Println("----------------------------")
		ui.Println("A wallet is already configured on this system.")
		ui.Println()
		ui.Println("What would you like to do?")
		ui.Println("  1. Skip wallet setup (use existing wallet)")
		ui.Println("  2. Replace existing wallet")
		ui.Println()

		choice, err := input.ReadInt("Enter choice [1-2]: ", 1, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to read choice: %w", err)
		}

		if choice == 1 {
			ui.Println("Using existing wallet...")
			return &WalletResult{
				Address: "existing_wallet",
				Created: false,
			}, nil
		}
		ui.Println()
	}

	ui.Println("🔐 Wallet Setup")
	ui.Println("---------------")
	ui.Println("Choose your setup method:")
	ui.Println("  1. Generate new mnemonic (recommended)")
	ui.Println("  2. Import existing mnemonic")
	ui.Println("  3. Import keystore JSON (MetaMask, etc.)")
	ui.Println("  4. Import private key")
	ui.Println()

	choice, err := input.ReadInt("Enter choice [1-4]: ", 1, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to read choice: %w", err)
	}

	ui.Println()
	password, err := input.ReadPassword("Password: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	score, label, _ := opts.PasswordStrength(password)
	ui.Printf("Password strength: %s (%d%%)\n", label, score)
	ui.Println()

	switch choice {
	case 1:
		return generateNewWallet(ui, input, wallet, opts, password)
	case 2:
		return importMnemonicWallet(ui, input, wallet, opts, password)
	case 3:
		return importKeystoreWallet(ui, input, wallet, opts, password)
	case 4:
		return importPrivateKeyWallet(ui, input, wallet, opts, password)
	default:
		return nil, fmt.Errorf("invalid choice: %d", choice)
	}
}

func generateNewWallet(ui UI, input *InputReader, wallet WalletOps, opts SetupOptions, password string) (*WalletResult, error) {
	ui.Println("Generating new mnemonic...")
	mnemonic, err := wallet.GenerateMnemonic()
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	copyErr := error(nil)
	if opts.CopyMnemonic != nil {
		copyErr = opts.CopyMnemonic(mnemonic)
	}

	ui.Println()
	ui.Println("⚠️  IMPORTANT: Save your mnemonic phrase!")
	ui.Println("Write these words down on paper and store them securely!")
	ui.Println()
	ui.Println(mnemonic)
	ui.Println()

	if opts.CopyMnemonic != nil {
		if copyErr == nil {
			ui.Println("✅ Mnemonic automatically copied to clipboard!")
		} else {
			ui.Println("⚠️  Failed to copy to clipboard, please copy manually!")
		}
	}

	ui.Println()
	ui.Println("Anyone with these words can access your funds!")
	ui.Println()
	_, _ = input.ReadLine("Press Enter after you've saved the mnemonic...")

	ui.Println()
	confirm, err := input.ReadLine("Have you saved the mnemonic? Type 'yes' to confirm: ")
	if err != nil {
		return nil, fmt.Errorf("failed to read confirmation: %w", err)
	}
	if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
		return nil, fmt.Errorf("mnemonic not confirmed")
	}

	address, err := wallet.CreateWallet(password, mnemonic, opts.WalletName)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &WalletResult{Address: address, Created: true}, nil
}

func importMnemonicWallet(ui UI, input *InputReader, wallet WalletOps, opts SetupOptions, password string) (*WalletResult, error) {
	ui.Println()
	ui.Println("Enter your 12 or 24 word mnemonic phrase:")
	ui.Println("(Input will be hidden for security)")
	mnemonic, err := input.ReadSecureLine("> ")
	if err != nil {
		return nil, fmt.Errorf("failed to read mnemonic: %w", err)
	}

	valid, errMsg := opts.ValidateMnemonic(mnemonic)
	if !valid {
		return nil, fmt.Errorf("invalid mnemonic: %s", errMsg)
	}

	address, err := wallet.CreateWallet(password, mnemonic, opts.WalletName)
	if err != nil {
		return nil, fmt.Errorf("failed to import mnemonic: %w", err)
	}

	return &WalletResult{Address: address, Created: true}, nil
}

func importKeystoreWallet(ui UI, input *InputReader, wallet WalletOps, opts SetupOptions, password string) (*WalletResult, error) {
	ui.Println()
	ui.Println("Choose import method:")
	ui.Println("  1. Paste keystore JSON content")
	ui.Println("  2. Enter file path to keystore")
	ui.Println()

	choice, err := input.ReadInt("Enter choice [1-2]: ", 1, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to read choice: %w", err)
	}

	var keystoreData string
	if choice == 1 {
		ui.Println()
		ui.Println("⚠️  Warning: Pasting keystore in terminal may be logged in shell history.")
		ui.Println("Consider using file import (option 2) for better security.")
		ui.Println()
		ui.Println("Paste your keystore JSON content:")
		keystoreData, err = input.ReadLine("> ")
		if err != nil {
			return nil, fmt.Errorf("failed to read keystore: %w", err)
		}
	} else {
		ui.Println()
		ui.Println("Enter the full path to your keystore file:")
		path, err := input.ReadLine("> ")
		if err != nil {
			return nil, fmt.Errorf("failed to read path: %w", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read keystore file: %w", err)
		}
		keystoreData = string(data)
	}

	valid, errMsg := opts.ValidateKeystore(keystoreData)
	if !valid {
		return nil, fmt.Errorf("invalid keystore: %s", errMsg)
	}

	address, err := wallet.ImportKeystore(keystoreData, password, opts.WalletName)
	if err != nil {
		return nil, fmt.Errorf("failed to import keystore: %w", err)
	}

	return &WalletResult{Address: address, Created: true}, nil
}

func importPrivateKeyWallet(ui UI, input *InputReader, wallet WalletOps, opts SetupOptions, password string) (*WalletResult, error) {
	ui.Println()
	ui.Println("⚠️  Security Warning: Never share your private key!")
	ui.Println("Enter your 64-character hex private key (input will be hidden):")
	ui.Println()

	privateKey, err := input.ReadSecureLine("Private key: ")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKey = strings.TrimPrefix(privateKey, "0X")

	valid, errMsg := opts.ValidatePrivKey(privateKey)
	if !valid {
		return nil, fmt.Errorf("invalid private key: %s", errMsg)
	}

	address, err := wallet.ImportPrivateKey(privateKey, password, opts.WalletName)
	if err != nil {
		return nil, fmt.Errorf("failed to import private key: %w", err)
	}

	return &WalletResult{Address: address, Created: true}, nil
}
