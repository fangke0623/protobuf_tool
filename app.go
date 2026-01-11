package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User struct represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // Hashed password
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Session struct represents a user session
type Session struct {
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// App struct
type App struct {
	ctx             context.Context
	users           map[string]User
	sessions        map[string]Session
	userDataFile    string
	sessionDataFile string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		users:    make(map[string]User),
		sessions: make(map[string]Session),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initUserData()
}

// initUserData initializes user data storage
func (a *App) initUserData() {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Set data file paths
	a.userDataFile = filepath.Join(appRoot, "data", "users.json")
	a.sessionDataFile = filepath.Join(appRoot, "data", "sessions.json")

	// Create data directory if it doesn't exist
	dataDir := filepath.Join(appRoot, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	// Load existing user data
	a.loadUserData()
	a.loadSessionData()
}

// loadUserData loads user data from file
func (a *App) loadUserData() {
	// Check if file exists
	if _, err := os.Stat(a.userDataFile); os.IsNotExist(err) {
		return // No data file yet
	}

	// Read file content
	content, err := os.ReadFile(a.userDataFile)
	if err != nil {
		fmt.Printf("Error reading user data file: %v\n", err)
		return
	}

	// Parse JSON data
	var users map[string]User
	if err := json.Unmarshal(content, &users); err != nil {
		fmt.Printf("Error parsing user data: %v\n", err)
		return
	}

	a.users = users
}

// saveUserData saves user data to file
func (a *App) saveUserData() {
	// Convert users map to JSON
	content, err := json.MarshalIndent(a.users, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling user data: %v\n", err)
		return
	}

	// Write to file
	if err := os.WriteFile(a.userDataFile, content, 0644); err != nil {
		fmt.Printf("Error writing user data file: %v\n", err)
	}
}

// loadSessionData loads session data from file
func (a *App) loadSessionData() {
	// Check if file exists
	if _, err := os.Stat(a.sessionDataFile); os.IsNotExist(err) {
		return // No data file yet
	}

	// Read file content
	content, err := os.ReadFile(a.sessionDataFile)
	if err != nil {
		fmt.Printf("Error reading session data file: %v\n", err)
		return
	}

	// Parse JSON data
	var sessions map[string]Session
	if err := json.Unmarshal(content, &sessions); err != nil {
		fmt.Printf("Error parsing session data: %v\n", err)
		return
	}

	a.sessions = sessions
}

// saveSessionData saves session data to file
func (a *App) saveSessionData() {
	// Convert sessions map to JSON
	content, err := json.MarshalIndent(a.sessions, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling session data: %v\n", err)
		return
	}

	// Write to file
	if err := os.WriteFile(a.sessionDataFile, content, 0644); err != nil {
		fmt.Printf("Error writing session data file: %v\n", err)
	}
}

// generateID generates a random ID for users and sessions
func (a *App) generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// ReadPB reads protobuf content from file
func (a *App) ReadPB(filename string) string {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error getting current directory: %v", err)
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Read protobuf content from file
	filePath := filepath.Join(appRoot, "pb", filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading file %s: %v", filePath, err)
	}

	return string(content)
}

// SavePB saves protobuf content to file
func (a *App) SavePB(filename, content string) string {
	// Get application root directory (where main.go is located)
	appRoot, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error getting current directory: %v", err)
	}

	// Ensure we're in the correct directory
	// Check if main.go exists in current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		// If main.go not found, try to find it in parent directories
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Create pb directory if not exists
	pbDir := filepath.Join(appRoot, "pb")
	if err := os.MkdirAll(pbDir, 0755); err != nil {
		return fmt.Sprintf("Error creating pb directory %s: %v", pbDir, err)
	}

	// Write protobuf content to file
	filePath := filepath.Join(pbDir, filename)

	// Check if we have write permission
	if _, err := os.Stat(pbDir); err != nil {
		return fmt.Sprintf("Error accessing pb directory %s: %v", pbDir, err)
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error saving file %s: %v", filePath, err)
	}

	// Verify file was created and get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf("File %s was not created successfully: %v", filePath, err)
	}

	// List files in pb directory to verify
	pbFiles, err := os.ReadDir(pbDir)
	if err != nil {
		return fmt.Sprintf("File saved successfully: %s\n"+
			"File size: %d bytes\n"+
			"Current directory: %s\n"+
			"Error listing pb directory: %v",
			filePath, fileInfo.Size(), appRoot, err)
	}

	// Create file list string
	fileList := "Files in pb directory:"
	for _, file := range pbFiles {
		fileList += fmt.Sprintf("\n  - %s", file.Name())
	}

	return fmt.Sprintf("File saved successfully: %s\n"+
		"File size: %d bytes\n"+
		"Current directory: %s\n"+
		"%s",
		filePath, fileInfo.Size(), appRoot, fileList)
}

// GenerateGRPC generates golang grpc code from protobuf
func (a *App) GenerateGRPC(filename, content string) string {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error getting current directory: %v", err)
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// First save the protobuf file
	pbDir := filepath.Join(appRoot, "pb")
	if err := os.MkdirAll(pbDir, 0755); err != nil {
		return fmt.Sprintf("Error creating pb directory %s: %v", pbDir, err)
	}

	filePath := filepath.Join(pbDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error saving file %s: %v", filePath, err)
	}

	// Create output directory
	outputDir := filepath.Join(appRoot, "grpc_output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Sprintf("Error creating output directory %s: %v", outputDir, err)
	}

	// Check if protoc is installed - first try project local protoc
	protocPath := filepath.Join(appRoot, "bin", "protoc")
	absProtocPath := protocPath

	// Check if local protoc exists
	if _, err := os.Stat(protocPath); os.IsNotExist(err) {
		// Try to find protoc in system paths
		if systemProtocPath, err := exec.LookPath("protoc"); err == nil {
			protocPath = systemProtocPath
			absProtocPath = systemProtocPath
		} else {
			return fmt.Sprintf("File saved successfully: %s\n"+
				"Note: protoc compiler not found. GRPC code generation skipped.\n"+
				"Local protoc: %s (not found)\n"+
				"To enable GRPC code generation, please install protoc and try again.\n"+
				"For macOS: brew install protobuf\n"+
				"For Ubuntu: apt install -y protobuf-compiler\n"+
				"For Windows: Download from https://github.com/protocolbuffers/protobuf/releases", filePath, absProtocPath)
		}
	}

	// Set explicit plugin paths
	protocGenGoPath := "$HOME/go/bin/protoc-gen-go"
	protocGenGoGrpcPath := "$HOME/go/bin/protoc-gen-go-grpc"

	// Expand environment variables
	protocGenGoPath = os.ExpandEnv(protocGenGoPath)
	protocGenGoGrpcPath = os.ExpandEnv(protocGenGoGrpcPath)

	// Check if plugins exist
	if _, err := os.Stat(protocGenGoPath); err != nil {
		return fmt.Sprintf("File saved successfully: %s\n"+
			"Warning: protoc-gen-go plugin not found at %s\n"+
			"Please install it with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest",
			filePath, protocGenGoPath)
	}

	if _, err := os.Stat(protocGenGoGrpcPath); err != nil {
		return fmt.Sprintf("File saved successfully: %s\n"+
			"Warning: protoc-gen-go-grpc plugin not found at %s\n"+
			"Please install it with: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
			filePath, protocGenGoGrpcPath)
	}

	// Create a temporary script to set PATH and run protoc
	tempScript := fmt.Sprintf(`#!/bin/bash
set -e

# Add go/bin to PATH
export PATH="$HOME/go/bin:$PATH"

# Set PROTOC_INCLUDE with all necessary paths
# 1. Local include directory
# 2. Google APIs from grpc-gateway
# 3. System protoc include directory (for descriptor.proto)
export PROTOC_INCLUDE="%s/include:/Users/fangke/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis:/opt/homebrew/include"

# Change to application root directory
cd "%s"

# Run protoc with the given arguments
"%s" --proto_path=".:${PROTOC_INCLUDE}" "$@"
`, appRoot, appRoot, protocPath)

	scriptPath := filepath.Join(appRoot, "run_protoc.sh")
	if err := os.WriteFile(scriptPath, []byte(tempScript), 0755); err != nil {
		return fmt.Sprintf("File saved successfully: %s\n"+
			"Warning: Error creating temporary script %s: %v", filePath, scriptPath, err)
	}
	defer os.Remove(scriptPath)

	// Debug information for protoc command - declared early for use in plugin checks
	debugInfo := fmt.Sprintf("\nDebug Information:\n"+
		"App Root: %s\n"+
		"PB File Path: %s\n"+
		"Output Directory: %s\n"+
		"Protoc Path: %s\n"+
		"Script Path: %s\n"+
		"Protoc Gen Go Path: %s\n"+
		"Protoc Gen Go GRPC Path: %s\n",
		appRoot, filePath, outputDir, protocPath, scriptPath, protocGenGoPath, protocGenGoGrpcPath)

	// Check for grpc-gateway plugin
	protocGenGrpcGatewayPath := "$HOME/go/bin/protoc-gen-grpc-gateway"
	protocGenGrpcGatewayPath = os.ExpandEnv(protocGenGrpcGatewayPath)
	grpcGatewayEnabled := true
	if _, err := os.Stat(protocGenGrpcGatewayPath); err != nil {
		grpcGatewayEnabled = false
		debugInfo += fmt.Sprintf("\ngRPC Gateway plugin not found at %s, skipping gateway generation", protocGenGrpcGatewayPath)
	}

	// Create simple test to verify output directory is writable
	testFile := filepath.Join(outputDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		debugInfo += fmt.Sprintf("\nError writing test file to output directory: %v", err)
	} else {
		debugInfo += fmt.Sprintf("\nSuccessfully wrote test file to output directory")
		os.Remove(testFile) // Clean up test file
	}

	// List output directory before generation
	beforeFiles, err := os.ReadDir(outputDir)
	if err == nil {
		debugInfo += "\nFiles in output directory before generation:"
		for _, file := range beforeFiles {
			debugInfo += fmt.Sprintf("\n  - %s", file.Name())
		}
	} else {
		debugInfo += fmt.Sprintf("\nError listing output directory before generation: %v", err)
	}

	// Get relative pb file path from app root
	relPbPath, err := filepath.Rel(appRoot, filePath)
	if err != nil {
		relPbPath = filepath.Base(filePath) // Fallback to just the filename
	}

	// Define all variables at function level
	var finalErr error
	var finalOutput string
	var finalCmdString string
	var cmdErr error
	var cmdErr2 error
	var cmdErr3 error

	// Get all proto files in pb directory
	pbFiles, err := filepath.Glob(filepath.Join(appRoot, "pb", "*.proto"))
	if err != nil {
		debugInfo += fmt.Sprintf("\nError finding proto files: %v", err)
		return fmt.Sprintf("Error finding proto files: %v\n%s", err, debugInfo)
	}

	// Convert to relative paths
	relPbFiles := make([]string, len(pbFiles))
	for i, file := range pbFiles {
		rel, err := filepath.Rel(appRoot, file)
		if err != nil {
			rel = filepath.Base(file)
		}
		relPbFiles[i] = rel
	}

	// Simplified approach: just use one command with source_relative for all proto files
	// This is the most reliable way for simple cases
	cmdArgs := []string{
		fmt.Sprintf("--go_out=paths=source_relative:%s", outputDir),
		fmt.Sprintf("--go-grpc_out=paths=source_relative:%s", outputDir),
	}
	cmdArgs = append(cmdArgs, relPbFiles...)

	debugInfo += fmt.Sprintf("\nTrying approach: Simple source_relative")

	// First approach
	cmd1 := exec.Command(scriptPath, cmdArgs...)
	output1, err1 := cmd1.CombinedOutput()
	cmdErr = err1
	finalOutput = string(output1)
	finalErr = cmdErr
	finalCmdString = cmd1.String()

	if cmdErr == nil {
		debugInfo += fmt.Sprintf("\n✅ Approach succeeded! Output: %s", finalOutput)
	} else {
		debugInfo += fmt.Sprintf("\n❌ Approach failed with error: %v\nOutput: %s", cmdErr, finalOutput)
	}

	// If error, try another approach with explicit proto_path
	if cmdErr != nil {
		debugInfo += fmt.Sprintf("\nTrying approach: With explicit proto_path")
		cmdArgs2 := []string{
			"--proto_path=.",
			fmt.Sprintf("--go_out=paths=source_relative:%s", outputDir),
			fmt.Sprintf("--go-grpc_out=paths=source_relative:%s", outputDir),
		}
		cmdArgs2 = append(cmdArgs2, relPbFiles...)

		cmd2 := exec.Command(scriptPath, cmdArgs2...)
		output2, err2 := cmd2.CombinedOutput()
		cmdErr2 = err2
		finalOutput = string(output2)
		finalErr = cmdErr2
		finalCmdString = cmd2.String()

		if cmdErr2 == nil {
			debugInfo += fmt.Sprintf("\n✅ Approach succeeded! Output: %s", finalOutput)
		} else {
			debugInfo += fmt.Sprintf("\n❌ Approach failed with error: %v\nOutput: %s", cmdErr2, finalOutput)
		}
	}

	// If still error, try with M option
	if finalErr != nil {
		debugInfo += fmt.Sprintf("\nTrying approach: With M option")
		cmdArgs3 := []string{
			fmt.Sprintf("--go_out=paths=source_relative:%s", outputDir),
			fmt.Sprintf("--go-grpc_out=paths=source_relative:%s", outputDir),
		}
		// Add M options for each proto file
		for _, file := range relPbFiles {
			cmdArgs3 = append(cmdArgs3, fmt.Sprintf("--go_opt=M%s=.", file))
			cmdArgs3 = append(cmdArgs3, fmt.Sprintf("--go-grpc_opt=M%s=.", file))
		}
		cmdArgs3 = append(cmdArgs3, relPbFiles...)

		cmd3 := exec.Command(scriptPath, cmdArgs3...)
		output3, err3 := cmd3.CombinedOutput()
		cmdErr3 = err3
		finalOutput = string(output3)
		finalErr = cmdErr3
		finalCmdString = cmd3.String()

		if cmdErr3 == nil {
			debugInfo += fmt.Sprintf("\n✅ Approach succeeded! Output: %s", finalOutput)
		} else {
			debugInfo += fmt.Sprintf("\n❌ Approach failed with error: %v\nOutput: %s", cmdErr3, finalOutput)
		}
	}

	if finalErr != nil {
		return fmt.Sprintf("File saved successfully: %s\n"+
			"Warning: Error generating GRPC code: %v\n"+
			"Output: %s\n"+
			"Command: %s\n"+
			"%s",
			filePath, finalErr, finalOutput, finalCmdString, debugInfo)
	}

	// List files in output directory to verify
	afterFiles, err := os.ReadDir(outputDir)
	fileList := ""
	if err == nil {
		fileList = "\nFiles in grpc_output directory after generation:"
		for _, file := range afterFiles {
			fileList += fmt.Sprintf("\n  - %s", file.Name())
			// Get file size
			if fileInfo, err := file.Info(); err == nil {
				fileList += fmt.Sprintf(" (size: %d bytes)", fileInfo.Size())
			}
		}
	} else {
		fileList = fmt.Sprintf("\nError listing output directory after generation: %v", err)
	}

	// Generate gRPC Gateway code if enabled
	if grpcGatewayEnabled {
		debugInfo += fmt.Sprintf("\n\n🔄 Generating gRPC Gateway code...")

		// Create gateway output directory
		gatewayOutputDir := outputDir

		// Define gateway command arguments
		gatewayCmdArgs := []string{
			fmt.Sprintf("--grpc-gateway_out=paths=source_relative:%s", gatewayOutputDir),
			"--grpc-gateway_opt=grpc_api_configuration=gateway.yaml",
			"--grpc-gateway_opt=allow_delete_body=true",
			relPbPath,
		}

		// Run gateway generation command
		gatewayCmd := exec.Command(scriptPath, gatewayCmdArgs...)
		gatewayOutput, gatewayErr := gatewayCmd.CombinedOutput()

		if gatewayErr == nil {
			debugInfo += fmt.Sprintf("\n✅ gRPC Gateway code generated successfully! Output: %s", string(gatewayOutput))

			// Update file list with gateway files
			afterGatewayFiles, _ := os.ReadDir(gatewayOutputDir)
			fileList += "\n\nFiles after gRPC Gateway generation:"
			for _, file := range afterGatewayFiles {
				fileList += fmt.Sprintf("\n  - %s", file.Name())
				// Get file size
				if fileInfo, err := file.Info(); err == nil {
					fileList += fmt.Sprintf(" (size: %d bytes)", fileInfo.Size())
				}
			}
		} else {
			debugInfo += fmt.Sprintf("\n⚠️  gRPC Gateway code generation failed: %v\nOutput: %s", gatewayErr, string(gatewayOutput))
			debugInfo += "\nNote: gRPC Gateway requires proper proto annotations and gateway.yaml configuration"
		}
	}

	return fmt.Sprintf("File saved successfully: %s\n"+
		"GRPC code generated successfully! Output saved to %s directory\n"+
		"Output: %s\n"+
		"%s%s", filePath, outputDir, finalOutput, debugInfo, fileList)
}

// GetGeneratedFiles returns a list of generated files
func (a *App) GetGeneratedFiles() []map[string]interface{} {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Get output directory - files are actually in grpc_output/pb
	outputDir := filepath.Join(appRoot, "grpc_output", "pb")

	// Read files in output directory
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil
	}

	// Create file list
	var fileList []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			fileList = append(fileList, map[string]interface{}{
				"name":     file.Name(),
				"size":     fileInfo.Size(),
				"modified": fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return fileList
}

// ReadGeneratedFile reads the content of a generated file
func (a *App) ReadGeneratedFile(filename string) string {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error getting current directory: %v", err)
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Get file path - files are actually in grpc_output/pb
	filePath := filepath.Join(appRoot, "grpc_output", "pb", filename)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading file %s: %v", filePath, err)
	}

	return string(content)
}

// DownloadGeneratedFile returns the content of a generated file for download
func (a *App) DownloadGeneratedFile(filename string) string {
	// This is the same as ReadGeneratedFile since we just need the file content
	return a.ReadGeneratedFile(filename)
}

// GetPBFiles returns a list of protobuf files in the pb directory
func (a *App) GetPBFiles() []map[string]interface{} {
	// Get application root directory
	appRoot, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Ensure we're in the correct directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		appRoot = "/Users/fangke/Documents/project/golang/pb-tool/pb-tool"
	}

	// Get pb directory
	pbDir := filepath.Join(appRoot, "pb")

	// Read files in pb directory
	files, err := os.ReadDir(pbDir)
	if err != nil {
		return nil
	}

	// Create file list
	var fileList []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".proto" {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			fileList = append(fileList, map[string]interface{}{
				"name":     file.Name(),
				"size":     fileInfo.Size(),
				"modified": fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return fileList
}

// RegisterUser registers a new user
func (a *App) RegisterUser(username, email, password string) string {
	// Check if username already exists
	for _, user := range a.users {
		if user.Username == username {
			return fmt.Sprintf(`{"error":"Username already exists"}`)
		}
		if user.Email == email {
			return fmt.Sprintf(`{"error":"Email already exists"}`)
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Sprintf(`{"error":"Failed to hash password: %v"}`, err)
	}

	// Create new user
	now := time.Now()
	userID := a.generateID()
	user := User{
		ID:        userID,
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user
	a.users[userID] = user
	a.saveUserData()

	// Generate session token
	token := a.generateID()
	session := Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(24 * time.Hour), // Session expires in 24 hours
	}

	// Save session
	a.sessions[token] = session
	a.saveSessionData()

	// Return success response
	return fmt.Sprintf(`{"success":true,"token":"%s","user":{"id":"%s","username":"%s","email":"%s"}}`, token, userID, username, email)
}

// LoginUser logs in a user
func (a *App) LoginUser(username, password string) string {
	// Find user by username
	var user User
	var found bool
	for _, u := range a.users {
		if u.Username == username {
			user = u
			found = true
			break
		}
	}

	if !found {
		return fmt.Sprintf(`{"error":"Invalid username or password"}`)
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return fmt.Sprintf(`{"error":"Invalid username or password"}`)
	}

	// Generate session token
	now := time.Now()
	token := a.generateID()
	session := Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: now.Add(24 * time.Hour), // Session expires in 24 hours
	}

	// Save session
	a.sessions[token] = session
	a.saveSessionData()

	// Return success response
	return fmt.Sprintf(`{"success":true,"token":"%s","user":{"id":"%s","username":"%s","email":"%s"}}`, token, user.ID, user.Username, user.Email)
}

// GetCurrentUser returns the current user based on session token
func (a *App) GetCurrentUser(token string) string {
	// Check if session exists and is valid
	session, ok := a.sessions[token]
	if !ok {
		return fmt.Sprintf(`{"error":"Invalid session"}`)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		delete(a.sessions, token)
		a.saveSessionData()
		return fmt.Sprintf(`{"error":"Session expired"}`)
	}

	// Get user
	user, ok := a.users[session.UserID]
	if !ok {
		return fmt.Sprintf(`{"error":"User not found"}`)
	}

	// Return user information
	return fmt.Sprintf(`{"success":true,"user":{"id":"%s","username":"%s","email":"%s"}}`, user.ID, user.Username, user.Email)
}

// LogoutUser logs out a user by removing their session
func (a *App) LogoutUser(token string) string {
	// Remove session
	delete(a.sessions, token)
	a.saveSessionData()

	return fmt.Sprintf(`{"success":true}`)
}
