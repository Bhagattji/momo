package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Enhanced self-update:
// - For `github` source: fetch release metadata, choose an asset (by name or OS), download artifact and its signature
// - Verify signature if GPG public key is provided (env GPG_PUBLIC_KEY or --pubkey file)
// - Replace current executable atomically on Unix-like systems; on Windows write .new and instruct user
// Usage examples:
//  momo self-update --source github --token $GITHUB_TOKEN --version v1.2.3
//  momo self-update --source url https://example.com/momo-linux --out /tmp/momo

func selfUpdateCmd(args []string) error {
	fs := flag.NewFlagSet("self-update", flag.ContinueOnError)
	source := fs.String("source", "github", "update source: github or url")
	token := fs.String("token", "", "auth token for private releases (if needed)")
	version := fs.String("version", "", "specific version tag to fetch (optional)")
	outPath := fs.String("out", "", "path to save downloaded artifact (optional)")
	assetName := fs.String("asset", "", "specific asset name to download from release (optional)")
	pubKeyFile := fs.String("pubkey", "", "path to GPG public key file to use for verification (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *source == "github" {
		repo := os.Getenv("GITHUB_REPO")
		if repo == "" {
			repo = "ORG/momo"
		}
		var metaURL string
		if *version == "" {
			metaURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
		} else {
			metaURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, *version)
		}
		req, _ := http.NewRequest("GET", metaURL, nil)
		if *token != "" {
			req.Header.Set("Authorization", "token "+*token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil { return err }
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed fetch release metadata: %s %s", resp.Status, string(b))
		}
		var meta map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
			return err
		}
		assets, _ := meta["assets"].([]interface{})
		if len(assets) == 0 {
			return fmt.Errorf("no assets found in release")
		}

		// choose asset: prefer explicit name, then runtime.GOOS, then first
		var chosen map[string]interface{}
		var sigAsset map[string]interface{}
		for _, a := range assets {
			m, ok := a.(map[string]interface{})
			if !ok { continue }
			name, _ := m["name"].(string)
			if *assetName != "" && name == *assetName {
				chosen = m
				break
			}
		}
		if chosen == nil {
			// try match by OS in name
			for _, a := range assets {
				m, ok := a.(map[string]interface{})
				if !ok { continue }
				name, _ := m["name"].(string)
				if strings.Contains(strings.ToLower(name), runtime.GOOS) || strings.Contains(strings.ToLower(name), "momo") {
					chosen = m
					break
				}
			}
		}
		if chosen == nil {
			// fallback to first asset
			if len(assets) > 0 {
				chosen, _ = assets[0].(map[string]interface{})
			}
		}
		if chosen == nil { return fmt.Errorf("failed to select asset") }

		// find signature asset (name endswith .sig or .asc or contains sig)
		for _, a := range assets {
			m, ok := a.(map[string]interface{})
			if !ok { continue }
			name, _ := m["name"].(string)
			if strings.HasSuffix(name, ".sig") || strings.HasSuffix(name, ".asc") || strings.Contains(strings.ToLower(name), "sig") {
				sigAsset = m
				break
			}
		}

		dl := func(url string, out string) error {
			req, _ := http.NewRequest("GET", url, nil)
			if *token != "" { req.Header.Set("Authorization", "token "+*token) }
			resp, err := http.DefaultClient.Do(req)
			if err != nil { return err }
			defer resp.Body.Close()
			if resp.StatusCode >= 400 { b, _ := io.ReadAll(resp.Body); return fmt.Errorf("download failed: %s %s", resp.Status, string(b)) }
			of, err := os.Create(out)
			if err != nil { return err }
			defer of.Close()
			_, err = io.Copy(of, resp.Body)
			return err
		}

		assetURL, _ := chosen["browser_download_url"].(string)
		assetNameStr, _ := chosen["name"].(string)
		out := *outPath
		if out == "" { out = filepath.Join(os.TempDir(), assetNameStr) }
		fmt.Println("Downloading asset to:", out)
		if err := dl(assetURL, out); err != nil { return err }

		var sigPath string
		if sigAsset != nil {
			sigURL, _ := sigAsset["browser_download_url"].(string)
			sigName, _ := sigAsset["name"].(string)
			sigPath = filepath.Join(os.TempDir(), sigName)
			fmt.Println("Downloading signature to:", sigPath)
			if err := dl(sigURL, sigPath); err != nil { return err }
		} else {
			fmt.Println("Warning: no signature asset found in release. Skipping verification unless pubkey provided with separate sig file.")
		}

		// If a public key is provided via flag or env, use it to verify
		pubKey := os.Getenv("GPG_PUBLIC_KEY")
		if *pubKeyFile != "" {
			b, err := os.ReadFile(*pubKeyFile)
			if err == nil { pubKey = string(b) }
		}

		if pubKey != "" && sigPath != "" {
			// create isolated GNUPGHOME
			gnupghome, err := os.MkdirTemp("", "momo-gnupg")
			if err != nil { return err }
			defer os.RemoveAll(gnupghome)
			os.Setenv("GNUPGHOME", gnupghome)
			pubf := filepath.Join(gnupghome, "pubkey.asc")
			if err := os.WriteFile(pubf, []byte(pubKey), 0600); err != nil { return err }
			// import
			cmd := exec.Command("gpg", "--batch", "--import", pubf)
			outb, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("gpg import failed: %v: %s", err, string(outb))
			}
			// verify
			cmd = exec.Command("gpg", "--batch", "--verify", sigPath, out)
			outb, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("gpg verify failed: %v: %s", err, string(outb))
			}
			fmt.Println("Signature verified via provided public key.")
		} else if pubKey != "" && sigPath == "" {
			return fmt.Errorf("public key supplied but no signature available to verify")
		} else {
			fmt.Println("No public key provided or no signature available; skipping verification.")
		}

		// Replace executable
		exePath, err := os.Executable()
		if err != nil { return err }
		fmt.Println("Current executable:", exePath)

		// Platform-specific replacement
		if runtime.GOOS == "windows" {
			// Writing new file next to exe and instruct user to restart / replace
			newPath := exePath + ".new"
			if err := os.WriteFile(newPath, nil, 0755); err != nil {
				// ignore; we'll attempt rename below
			}
			fmt.Println("On Windows, automatic replacement may fail while running. New binary saved at:", out)
			fmt.Println("To complete update: stop the running process and replace", exePath, "with", out)
			return nil
		}

		backup := exePath + ".bak." + time.Now().Format("20060102T150405")
		if err := os.Rename(exePath, backup); err != nil {
			return fmt.Errorf("failed to backup current executable: %v", err)
		}
		if err := os.Rename(out, exePath); err != nil {
			// attempt rollback
			_ = os.Rename(backup, exePath)
			return fmt.Errorf("failed to replace executable: %v", err)
		}
		fmt.Println("Update applied. Backup of previous binary:", backup)
		return nil
	}

	// source == url: download direct URL and save
	if *source == "url" {
		u := fs.Arg(0)
		if u == "" { return fmt.Errorf("url source requires a URL argument") }
		out := *outPath
		if out == "" { out = filepath.Join(os.TempDir(), filepath.Base(u)) }
		resp, err := http.Get(u)
		if err != nil { return err }
		defer resp.Body.Close()
		if resp.StatusCode >= 400 { return fmt.Errorf("failed download: %s", resp.Status) }
		of, err := os.Create(out)
		if err != nil { return err }
		defer of.Close()
		if _, err := io.Copy(of, resp.Body); err != nil { return err }
		fmt.Println("Downloaded artifact to:", out)
		fmt.Println("Note: verify signature manually before replacing binary")
		return nil
	}

	return fmt.Errorf("unknown source: %s", *source)
}
