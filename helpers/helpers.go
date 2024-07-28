package helpers

import (
	"fmt"
	"go-getpi/client"
	"go-getpi/config"
	"log"
	"os"
)

func InitLogger() {
	logFile, err := os.OpenFile("application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func SyncPihole(cfg *config.Config, updateGravity bool, interval int, hosts []string, sslSecure bool) error {
	fmt.Println("Starting SyncPihole process...")
	log.Println("Starting SyncPihole process...")

	primaryHost, err := initializePrimaryHost(cfg)
	if err != nil {
		return err
	}

	backupFile, err := performBackup(primaryHost)
	if err != nil {
		return err
	}
	if backupFile != "" {
		err = uploadBackupToSecondaryHosts(cfg, backupFile, sslSecure)
		if err != nil {
			return err
		}
	} else {
		log.Println("Backup file not found, skipping upload to secondary hosts.")
		return fmt.Errorf("backup file not found")
	}

	if updateGravity {
		if err := updateGravityProcess(primaryHost); err != nil {
			return err
		}
		for _, secondaryHost := range cfg.SecondaryHosts {
			if err := updateGravityOnSecondaryHost(secondaryHost); err != nil {
				log.Printf("Failed to update gravity on secondary host %s: %v", secondaryHost.BaseURL, err)
			} else {
				fmt.Printf("Gravity updated successfully on secondary host: %s\n", secondaryHost.BaseURL)
				log.Printf("Gravity updated successfully on secondary host: %s\n", secondaryHost.BaseURL)
			}
		}
	}

	return nil
}

func initializePrimaryHost(cfg *config.Config) (*client.Client, error) {
	primaryHost := client.NewHost(cfg.PrimaryHost.BaseURL, cfg.PrimaryHost.Password, cfg.PrimaryHost.Path)
	primaryHost.SslSecure = cfg.PrimaryHost.SslSecure
	fmt.Printf("Primary host: %s\n", primaryHost.FullURL)
	log.Printf("Primary host: %s\n", primaryHost.FullURL)

	piholeClient, err := client.NewClient(*primaryHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	fmt.Println("Logging in to primary host...")
	log.Println("Logging in to primary host...")
	token, err := piholeClient.Login(false)
	if err != nil {
		return nil, fmt.Errorf("failed to login to primary host: %v", err)
	}
	fmt.Printf("Login successful, token: %s\n", token)
	log.Printf("Login successful, token: %s\n", token)
	return piholeClient, nil
}

func performBackup(primaryHost *client.Client) (string, error) {
	fmt.Println("Starting backup process...")
	log.Println("Starting backup process...")
	backupData, err := primaryHost.DownloadBackup()
	if err != nil {
		return "", fmt.Errorf("failed to download backup: %v", err)
	}
	fmt.Println("Backup downloaded successfully.")
	log.Println("Backup downloaded successfully.")

	backupFile := "backup.gz"
	err = os.WriteFile(backupFile, backupData, 0644)
	if err != nil {
		return "", fmt.Errorf("error saving backup file: %v", err)
	}
	fmt.Println("Backup saved to file:", backupFile)
	log.Println("Backup saved to file:", backupFile)

	return backupFile, nil
}

func uploadBackupToSecondaryHosts(cfg *config.Config, backupFile string, sslSecure bool) error {
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}

	for _, secondaryHost := range cfg.SecondaryHosts {
		if err := uploadBackupToHost(secondaryHost, backupData, sslSecure); err != nil {
			log.Printf("Failed to upload backup to secondary host %s: %v", secondaryHost.BaseURL, err)
		} else {
			fmt.Printf("Backup uploaded successfully to secondary host: %s\n", secondaryHost.BaseURL)
			log.Printf("Backup uploaded successfully to secondary host: %s\n", secondaryHost.BaseURL)
		}
	}

	return nil
}

func uploadBackupToHost(host config.Host, backupData []byte, sslSecure bool) error {
	hostConfig := client.NewHost(host.BaseURL, host.Password, "/admin/")
	hostConfig.SslSecure = sslSecure
	secondaryClient, err := client.NewClient(*hostConfig)
	if err != nil {
		return fmt.Errorf("failed to create secondary client: %v", err)
	}
	fmt.Printf("Logging in to secondary host: %s\n", hostConfig.FullURL)
	log.Printf("Logging in to secondary host: %s\n", hostConfig.FullURL)
	secondaryToken, err := secondaryClient.Login(false)
	if err != nil {
		return fmt.Errorf("failed to login to secondary host: %v", err)
	}
	secondaryClient.Token = secondaryToken

	fmt.Printf("Uploading backup to secondary host: %s\n", hostConfig.FullURL)
	log.Printf("Uploading backup to secondary host: %s\n", hostConfig.FullURL)
	success, err := secondaryClient.UploadBackup(backupData)
	if err != nil || !success {
		return fmt.Errorf("failed to upload backup to secondary host: %v", err)
	}
	return nil
}

func updateGravityProcess(primaryHost *client.Client) error {
	fmt.Println("Updating gravity on primary host...")
	log.Println("Updating gravity on primary host...")
	success, err := primaryHost.UpdateGravity()
	if err != nil {
		return fmt.Errorf("failed to update gravity on primary host: %v", err)
	}
	if !success {
		return fmt.Errorf("gravity update on primary host was not successful")
	}
	fmt.Println("Gravity update successful.")
	log.Println("Gravity update successful.")
	return nil
}

func updateGravityOnSecondaryHost(host config.Host) error {
	hostConfig := client.NewHost(host.BaseURL, host.Password, "/admin/")
	hostConfig.SslSecure = host.SslSecure
	secondaryClient, err := client.NewClient(*hostConfig)
	if err != nil {
		return fmt.Errorf("failed to create secondary client: %v", err)
	}
	fmt.Printf("Logging in to secondary host: %s\n", hostConfig.FullURL)
	log.Printf("Logging in to secondary host: %s\n", hostConfig.FullURL)
	secondaryToken, err := secondaryClient.Login(false)
	if err != nil {
		return fmt.Errorf("failed to login to secondary host: %v", err)
	}
	secondaryClient.Token = secondaryToken

	fmt.Printf("Updating gravity on secondary host: %s\n", hostConfig.FullURL)
	log.Printf("Updating gravity on secondary host: %s\n", hostConfig.FullURL)
	success, err := secondaryClient.UpdateGravity()
	if err != nil || !success {
		return fmt.Errorf("failed to update gravity on secondary host: %v", err)
	}
	return nil
}
