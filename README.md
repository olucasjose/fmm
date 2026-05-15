# FMM (Fastest Mint Mirrors)

🌍 *Read this in other languages: [English](README.md) | [Português](README.pt-BR.md)*

FMM is a fast and interactive mirror ranking and selection tool for Linux Mint. It is designed to quickly benchmark available Linux Mint mirrors and help you update your system's mirror configuration for optimal download speeds.

## Scope & Philosophy

**FMM focuses exclusively on testing and updating mirrors.**

Other features found in the original `mintsources` tool—such as GPG key management, PPA management, and other repository configurations—are **not** part of this project's scope. These features have not been implemented, and there are no plans to implement them in the future. FMM is built to do one thing and do it extremely well: find and set the fastest mirrors.

## Lineage & Credits

FMM uses the original [mintsources](https://github.com/linuxmint/mintsources) project as its direct base. This project is a Go-based port and derivative work of the `mintsources` mirror testing logic, developed and maintained by the Linux Mint team. FMM builds upon this robust foundation to provide a faster, CLI-driven experience.

FMM is licensed under the **GNU General Public License v3.0 (GPLv3)**. The original `mintsources` software is distributed under the GNU GPL. See the [LICENSE](LICENSE) file for more details.

## Installation

You can install FMM directly by running the provided installation script. The script compiles the Go binary, moves it to your system path, and sets up bash completion.

```bash
# Clone the repository
git clone https://github.com/olucasjose/fmm.git
cd fmm

# Run the installation script (requires sudo privileges)
./install.sh
```

## Usage

FMM requires `sudo` privileges for actions that modify system configurations (like `run`), but informational commands can be run as a regular user.

- **Run the main mirror ranking and update flow:**
  ```bash
  sudo fmm run
  ```

- **List available mirrors:**
  ```bash
  fmm list
  ```

- **Show the current mirror ranking (benchmark without applying):**
  ```bash
  fmm ranking
  ```

- **Show help and available commands:**
  ```bash
  fmm help
  ```
