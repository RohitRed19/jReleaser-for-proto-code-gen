package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// MavenProject represents the structure of the POM file for parsing XML
type MavenProject struct {
	XMLName xml.Name `xml:"project"`
	Version string   `xml:"version"`
}

// Clean all files except `go.mod` in each `go` subdirectory
func cleanGeneratedFiles(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip `go.mod` files
		if strings.HasSuffix(info.Name(), "go.mod") {
			return nil
		}

		// Skip `go.mod` files
		if strings.HasSuffix(info.Name(), "go.sum") {
			return nil
		}

		// Delete all other files
		fmt.Printf("Removing `%s`\n", path)
		removeErr := os.Remove(path)
		if removeErr != nil {
			return fmt.Errorf("failed to remove file '%s': %w", path, removeErr)
		}
		return nil
	})

	return err
}

func upgradeGoMod(dir string) {
	fmt.Printf("Upgrading Go module in `%s`\n", dir)

	// Step 1: Extract the major version from `pom.xml`
	majorVersion, err := extractMajorVersionFromPom("pom.xml")
	if err != nil {
		log.Fatalf("failed to extract major version from POM: %v", err)
	}

	// Step 2: Move existing files to the versioned directory
	currentVersionedDirectory, err := moveFilesToVersionedDirectory(dir, majorVersion)
	if err != nil {
		log.Fatalf("failed to move files to versioned directory: %v", err)
	}

	// Step 3: Check if `go.mod` exists.
	goModPath := filepath.Join(currentVersionedDirectory, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Println("`go.mod` does not exist. Initializing a new Go module.")

		// Extract a module name from the directory (default fallback).
		moduleName := "github.com/RohitRed19/jReleaser-for-proto-code-gen/" + filepath.Base(filepath.Clean(dir)) + "/go/v" + majorVersion

		// Run `go mod init <moduleName>`
		_, _, err := runCommand(currentVersionedDirectory, "go", "mod", "init", moduleName)
		if err != nil {
			log.Fatalf("Failed to initialize Go module: %v", err)
		}
		fmt.Printf("Initialized Go module with name `%s`.\n", moduleName)
	} else {
		fmt.Println("`go.mod` exists. Proceeding with upgrades.")
	}

	// Step 4: Get the current module name from `go.mod`
	moduleName, _, err := runCommand(currentVersionedDirectory, "go", "list", "-m")
	if err != nil {
		log.Fatalf("Error getting module name: %v", err)
	}

	// Step 5: Update the module path with the new major version
	newModuleName, err := changeModulePathVersion(moduleName, "v"+majorVersion)
	if err != nil {
		log.Fatalf("Error updating module path: %v", err)
	}

	if newModuleName == moduleName {
		fmt.Printf("Module name `%s` is already up-to-date.\n", newModuleName)
		return
	}

	// Step 6: Updates module name in `go.mod`
	_, _, err = runCommand(currentVersionedDirectory, "go", "mod", "edit", "-module", newModuleName)
	if err != nil {
		log.Fatalf("Failed to edit module name: %v", err)
	}
	fmt.Printf("Successfully updated module name to `%s`.\n", newModuleName)
}

func moveFilesToVersionedDirectory(baseDir string, majorVersion string) (string, error) {
	// Step 1: Construct the concerned directories for current and previous versions
	currentVersionedDir := filepath.Join(baseDir, "go", "v"+majorVersion)

	// Convert major version to int for calculating the previous version directory
	majorVersionInt, err := strconv.Atoi(majorVersion)
	if err != nil {
		return "", fmt.Errorf("major version %s is not a valid number: %w", majorVersion, err)
	}

	previousVersionedDir := ""
	if majorVersionInt > 1 {
		previousVersionedDir = filepath.Join(baseDir, "go", "v"+strconv.Itoa(majorVersionInt-1))
	}

	// Step 2: If no previous version exists, move unversioned files into the current directory
	if previousVersionedDir == "" || !directoryExists(previousVersionedDir) {
		// Create the current versioned directory if it doesn't exist
		if err := os.MkdirAll(currentVersionedDir, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", currentVersionedDir, err)
		}

		// Move unversioned files from `baseDir/go` into `currentVersionedDir`
		unversionedDir := filepath.Join(baseDir, "go")
		return currentVersionedDir, moveFiles(unversionedDir, currentVersionedDir)
	}

	// Step 3: If the previous version directory exists, move files from the previous directory
	if directoryExists(previousVersionedDir) {
		// Create the current versioned directory if it doesn't exist
		if err := os.MkdirAll(currentVersionedDir, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", currentVersionedDir, err)
		}

		// Move files from the previous version directory to the current one
		return currentVersionedDir, moveFiles(previousVersionedDir, currentVersionedDir)
	}

	return currentVersionedDir, nil
}

// Helper function to check if a directory exists
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Helper function to move all files from one directory to another
func moveFiles(srcDir, destDir string) error {
	// Walk through the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through %s: %w", srcDir, err)
		}

		// Skip the destination directory to prevent duplication
		if path == destDir {
			return filepath.SkipDir
		}

		// Skip directories (we only want to move files)
		if info.IsDir() {
			return nil
		}

		// Compute the relative path of the file
		relPath, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return fmt.Errorf("failed to compute relative file path: %w", relErr)
		}

		// Compute the destination path for the file
		destPath := filepath.Join(destDir, relPath)

		// Ensure the destination directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Move the file
		if err := os.Rename(path, destPath); err != nil {
			return fmt.Errorf("failed to move file %s to %s: %w", path, destPath, err)
		}

		fmt.Printf("Moved `%s` -> `%s`\n", path, destPath)
		return nil
	})
}

func runCommand(dir string, command string, args ...string) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Dir = filepath.Join(dir)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), err
}

func changeModulePathVersion(moduleName string, newVersion string) (string, error) {
	parts := strings.Split(moduleName, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("module name is empty or invalid")
	}

	// Regular expression to check if the last part is a version (e.g., v1, v2, v3, etc.)
	versionRegex := regexp.MustCompile(`^v\d+$`)

	// Check if the last part is a version (v1, v2, etc.)
	lastPart := parts[len(parts)-1]
	if versionRegex.MatchString(lastPart) {
		// If the last part is a version, replace it with the new version
		parts[len(parts)-1] = newVersion
	} else {
		// If the last part is not a version, append the new version
		parts = append(parts, newVersion)
	}

	return strings.Join(parts, "/"), nil
}

// Helper function to extract the major version from a POM file
func extractMajorVersionFromPom(pomPath string) (string, error) {
	// Open the pom.xml file
	file, err := os.Open(pomPath)
	if err != nil {
		return "", fmt.Errorf("could not open file %s: %v", pomPath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Parse the pom.xml file
	var project MavenProject
	decoder := xml.NewDecoder(file)
	err = decoder.Decode(&project)
	if err != nil {
		return "", fmt.Errorf("failed to parse pom.xml: %v", err)
	}

	// Ensure the version tag exists
	if strings.TrimSpace(project.Version) == "" {
		return "", fmt.Errorf("no <version> tag found in POM file")
	}

	// Extract the major version using regex
	re := regexp.MustCompile(`^(\d+)\.`)
	matches := re.FindStringSubmatch(project.Version)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("no major version found in version: %s", project.Version)
}

// Entrypoint
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [clean|upgrade-mod] <directory>\n", os.Args[0])
	}

	command := os.Args[1]
	if len(os.Args) < 3 {
		log.Fatalf("Please specify a directory to operate on.")
	}

	dir := os.Args[2]

	switch command {
	case "clean":
		fmt.Printf("Cleaning generated files in: %s\n", dir)
		err := cleanGeneratedFiles(dir)
		if err != nil {
			log.Fatalf("Failed to clean files: %v\n", err)
		}
		fmt.Println("Clean completed successfully.")
	case "upgrade-mod":
		upgradeGoMod(dir)
	default:
		log.Fatalf("Unknown command: %s. Use 'clean' or 'upgrade-mod'.\n", command)
	}
}
