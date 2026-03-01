package update

import (
	"fmt"
	"os"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/pkg/update"
)

func runUpdate(checkOnly bool) error {
	fmt.Printf("%s Checking for updates...\n", internal.Logo)

	release, err := update.CheckLatest()
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	latestVersion := release.TagName
	currentVersion := internal.GetVersion()

	if !update.IsNewer(currentVersion, latestVersion) {
		fmt.Printf("Already up to date! (current: %s, latest: %s)\n", currentVersion, latestVersion)
		return nil
	}

	fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
	fmt.Printf("Release: %s\n", release.HTMLURL)

	if checkOnly {
		return nil
	}

	assetURL, err := update.FindAssetURL(release)
	if err != nil {
		fmt.Printf("You can download manually from: %s\n", release.HTMLURL)
		return fmt.Errorf("finding asset: %w", err)
	}

	fmt.Printf("\nUpdate to %s? (y/n): ", latestVersion)
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("Update cancelled.")
		return nil
	}

	if err := update.DownloadAndReplace(assetURL, os.Stdout); err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	fmt.Printf("\n%s Updated to %s!\n", internal.Logo, latestVersion)
	return nil
}
