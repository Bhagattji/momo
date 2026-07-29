# momo — AI Coding Agent

## Install (Windows PowerShell)

`powershell
irm https://raw.githubusercontent.com/Bhagattji/momo/main/scripts/install.ps1 | iex
Verify Signature
gpg --import pubkey.asc
gpg --verify momo-windows-amd64.exe.sig momo-windows-amd64.exe
Usage
momo                    # Interactive TUI
momo --version          # Show version
momo --provider groq    # Use specific provider
momo --auto             # Auto-approve tools
momo self-update        # Update to latest release
Features
- Multi-provider support (Groq, OpenAI, Ollama, Anthropic)
- Agent loop with tool calling
- TUI permission dialogs
- Secure self-update with GPG verification
- Cross-platform (Windows, Linux, macOS)
- Signed releases with SBOM
