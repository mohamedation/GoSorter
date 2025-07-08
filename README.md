# GoSorter

A high-performance command-line file organizer written in Go that sorts files into folders based on their extensions.

## Features

- **Extension-based sorting**: Automatically organizes files into folders based on their file extensions
- **Duplicate detection**: Find and move duplicate files to a separate folder using hash comparison
- **Transparent PNG detection**: Special handling for PNG files with transparent backgrounds
- **Customizable configuration**: Define custom extension-to-folder mappings
- **Performance optimized**: Multi-threaded processing for large directories
- **Detailed statistics**: Comprehensive reporting on processed files


## About This Project

This is a personal project born out of frustration with my messy Downloads folder and scattered files across my system (ahem, systems). 

GoSorter is yet another cli tool that automates this process while serving as a practical way for me to learn and improve my Go programming skills. It might not be suitable for everybody or production just yet.

I am also working on a GUI for it GuiSorter.


## Installation

```bash
go install github.com/mohamedation/GoSorter@latest
```

Or build from source:

```bash
git clone https://github.com/mohamedation/GoSorter.git
cd GoSorter
go build -o gosorter
```

## Using Make

Alternatively, you can use `make` to build the project:

```bash
# Build the application locally
make build

# Run from current directory
./gosorter
```
This will compile the binary and place it in the current directory.

### System Install (Available Globally)
```bash
# Clone the repository
git clone https://github.com/mohamedation/GoSorter.git
cd GoSorter

# Install system-wide (requires sudo)
make install

# Now you can use gosorter from anywhere
gosorter /path/to/directory
```

### Uninstall
```bash
# Remove from system
make uninstall
```


## Usage

```bash
# Sort files in current directory
./gosorter

# Sort files in a specific directory
./gosorter /path/to/directory

# Enable duplicate detection (default 1GB max hash size)
./gosorter -d /path/to/directory

# Set max file size for hashing to 2GB
./gosorter -d -S 2G /path/to/directory

# Set max file size for hashing to 2048MB
./gosorter -d -S 2048M /path/to/directory

# Only detect and move duplicates (no extension-based sorting)
./gosorter -do /path/to/directory

# Enable verbose output
./gosorter -v /path/to/directory

# Enable silent mode (errors only)
./gosorter -s /path/to/directory

# Enable transparent PNG detection
./gosorter -t /path/to/directory

# Enable logging to file
./gosorter -l /path/to/directory

# Combine options
./gosorter -d -v -t /path/to/directory
```

## Options

- `-h`: Show help message
- `-d`: Move duplicate files to the Duplicates folder
- `-do`: Only detect and move duplicates, no extension-based sorting
- `-v`: Enable verbose output with detailed statistics
- `-s`: Enable silent mode (only show errors)
- `-l`: Enable logging to a file in the current directory
- `-t`: Check transparent PNGs (slower, but sorts PNGs with transparent backgrounds)
- `-S <size>`: Set maximum file size for hashing (e.g., `-S 2G` or `-S 2048M`)

## File Organization

The program organizes files into the following folders:

### Images
- **Pictures**: `.jpg`, `.jpeg`, `.png` (non-transparent), `.bmp`, `.heic`, `.heif`, `.tiff`, `.tif`
- **PNGs**: `.png` (with transparency)
- **GIFs**: `.gif`
- **SVGs**: `.svg`
- **WebP**: `.webp`
- **RawImages**: `.raw`

### Media
- **Videos**: `.mp4`, `.mkv`, `.avi`, `.mpg`, `.mpeg`, `.webm`
- **Music**: `.mp3`, `.wav`, `.flac`, `.aac`, `.ogg`, `.m4a`, `.wma`, `.opus`, `.m4b`, `.m4p`

### Documents & Office
- **Documents**: `.txt`, `.doc`, `.docx`, `.odt`, `.epub`
- **PDFs**: `.pdf`
- **Presentations**: `.ppt`, `.pptx`, `.odp`
- **Sheets**: `.csv`, `.xls`, `.xlsx`, `.ods`

### Design & Creative
- **Photoshop**: `.psd`
- **Illustrator**: `.ai`
- **InDesign**: `.indd`

### Archives & System
- **Archives**: `.zip`, `.rar`, `.tar`, `.gz`
- **Archives-Extracted**: `.zip` files with matching extracted folders
- **DiskImages**: `.dmg`
- **ISOs**: `.iso`

### Development & Data
- **Development**: `.py`
- **JSONs**: `.json`
- **XMLs**: `.xml`

### Applications & Executables
- **Executables**: `.exe`
- **AndroidApps**: `.apk`
- **Packages**: `.deb`

### Virtual Machines & Configs
- **VirtualMachines**: `.ova`
- **Configs**: `.ovpn`

### Other
- **3D**: `.stl`, `.3mf`, `.obj`
- **Torrents**: `.torrent`
- **Duplicates**: Duplicate files (when `-d` flag is used)

> **Note**: The list of supported extensions is not exhaustive. As I encounter new file types in my workflow, I add them incrementally. The configuration file allows anyone to customize or expand the folder and extension mappings to suit their needs, so you can easily adapt GoSorter to your own file organization preferences.

For the complete and up-to-date list of all supported extensions and their folder mappings, see the [extension configuration model](model/extension_config.go). 


## Configuration

Custom extension mappings can be configured in:
- **macOS/Linux**: `~/.config/GoSorter/extension.json`
- **Windows**: `%USERPROFILE%\.config\GoSorter\extension.json` (untested)

Example configuration:

```json
{
  "extension_to_folder": {
    ".jpg": "MyPhotos",
    ".pdf": "Documents",
    ".custom": "CustomFolder"
  },
  "archives_extracted_folder": "Archives-Extracted",
  "duplicates_folder": "Duplicates",
  "transparent_png_folder": "PNGs"
}
```
### Configuration Options

- **`extension_to_folder`**: Map file extensions to folder names
- **`archives_extracted_folder`**: Folder name for extracted archive contents
- **`duplicates_folder`**: Folder name for duplicate files (when using `-d` flag)
- **`transparent_png_folder`**: Folder name for transparent PNG files (when using `-t` flag)

**Example**: If you only specify `".jpg": "MyPhotos"` in your config, all other extensions will use their default mappings, but `.jpg` files will go to the "MyPhotos" folder instead of "Pictures".


## Package Usage

GoSorter can also be used as a Go package in other applications like my GuiSorter:

```go
import (
    "github.com/mohamedation/GoSorter/service"
    "github.com/mohamedation/GoSorter/model"
    "github.com/mohamedation/GoSorter/helpers"
)

// Create configuration
config := &model.Config{
    MoveDuplicates: true,
    Verbose: true,
}

// Create stats tracker
stats := &model.Stats{}

// Create file processor
processor := service.NewFileProcessor(config, stats, &helpers.CLILogger{})

// Process directory
err := processor.ProcessDirectory(context.Background(), "/path/to/directory")
```

## Development

If you find this project interesting, you are welcome to contribute, test, or use it in your own projects. I would appreciate it if you shared your awesome projects with me! I am striving to make GoSorter user-friendly and to follow best practices (from my point of view and understanding). If something doesn’t work, please let me know and I will work on fixing it.


### Building

```bash
# Format code
make fmt

# Run linting
make vet

# Run tests
make test

# Run tests with coverage
make coverage

# Build locally (development)
make build

# Build with quality checks
make build-local

# Install system-wide
make install

# Build for production with optimizations
make release

# Build for all platforms (local)
make release-all

# Cross-compile for specific platforms
make build-linux
make build-macos
make build-windows
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Platform Compatibility

- **Tested on**: macOS Intel, Linux (Ubuntu/Debian and also Arch)
- **Expected to work**: Windows, macOS Apple Silicon (tests to be preformed, You are welcome to test)
- **Go Version**: 1.23.4+ (Probably works with previous versions as well)

## Disclaimer

This software is under active development. While functional, it should be used with caution:

- **Always backup your files** before running the sorter
- **Test in a safe environment** first (use a copy of your files)
- **Report any issues** via GitHub issues
- **Use at your own risk** - the authors are not responsible for data loss


## License

GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

This software is free and open source. You can use, modify, and distribute it under the terms of the GPL v3, which ensures it remains free and open source forever.