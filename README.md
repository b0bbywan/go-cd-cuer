# go-disc-cuer

`go-disc-cuer` is a CLI tool for generating CUE files from audio CDs, with support for metadata enrichment from GNDB and MusicBrainz. It leverages `libdiscid` for disc ID calculation and provides options for customization and integration into workflows.

## Features

- **Disc ID Calculation**: Uses `libdiscid` to compute MusicBrainz and GNUDB compatible disc IDs.
- **Metadata Integration**: Fetch track and album metadata from GNUDB or MusicBrainz.
- **Fix Incorrect CUE Files**: Force the use of a specific MusicBrainz release to correct or regenerate CUE files.
- **Configurable**: Allows configuration through files, environment variables, and command-line flags.

## Installation

### From Source
1. Clone the repository:
    ```bash
    git clone https://github.com/b0bbywan/go-disc-cuer.git
    cd go-disc-cuer
    ```
2. Install dependencies
    ```bash
    # Debian
    sudo apt install libdiscid0 libdiscid-dev
    # Fedora
    sudo dnf install libdiscid libdiscid-devel
    ```

3. Build the binary:
    ```bash
    go build -o disc-cuer .
    ```

4. (Optional) Install globally:
    ```bash
    sudo mv disc-cuer /usr/local/bin/
    ```

## Usage
1. Basic Command
Generate a CUE file for the current CD:
    ```bash
    disc-cuer
    ```

2. Options
`--overwrite`: Regenerate the CUE file even if it exists.
`--musicbrainz` <release_id>: Specify a MusicBrainz release ID to fetch album metadata.
`--disc-id` <disc_id>: Provide a custom disc ID. This requires --musicbrainz to associate metadata with the ID.

3. Configuration
The tool loads configurations in the following order of priority:

4. Command-line flags.
- Environment variables prefixed with DISC_CUER_.
- Configuration files located at:
   - `/etc/disc-cuer/config.yml`
   - `~/.config/disc-cuer/config.yaml`

5. Example Configuration
**Please note that gnuHelloEmail is mandatory to use gnudb source**
    ```yaml
    gnuHelloEmail: "your-email@example.com"
    gnuDbUrl: "https://gnudb.gnudb.org"
    cacheLocation: "/var/cache/disc-cuer"
    ```

## Examples
- Fetch Metadata from music brainz and Generate CUE for current disc
    ```bash
    disc-cuer --musicbrainz <release_id> --overwrite
    ```
- Use Custom Disc ID
    ```bash
    disc-cuer --disc-id <disc_id> --musicbrainz <release_id> --overwrite
    ```

## Project Structure
`main/`: Entry point and CLI logic.
`cue/`: CUE file generation and related utilities.
`discinfo/`: Disc ID and metadata fetching logic.
`gnudb/`: GNDB integration.
`musicbrainz/`: MusicBrainz integration.
`config`: Configuration package with github.com/spf13/viper.
`utils/`: Shared helper functions.


## Contributing
Contributions are welcome! Feel free to submit issues, feature requests, or pull requests.

- Fork the repository.
- Create a feature branch.
- Commit your changes and open a pull request.