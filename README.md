# playtime

Personal retro games library + [EmulatorJS](https://emulatorjs.org/) + Netplay.

Features:

+ Multiple users with disk usage quota
+ Keyboard and gamepad controls
+ Save states
+ Netplay - up to 4 players

Platforms supported:

+ 3DO
+ Arcade
+ Atari 2600
+ Atari 5200
+ Atari 7800
+ Atari Jaguar
+ Atari Lynx
+ Bandai WonderSwan (Color)
+ ColecoVision
+ Commodore 128
+ Commodore 64
+ Commodore Amiga
+ Commodore PET
+ Commodore Plus/4
+ Commodore VIC-20
+ MAME 2003
+ NEC PC-FX
+ NEC TurboGrafx-16 / SuperGrafx / PC Engine
+ Nintendo 64
+ Nintendo DS
+ Nintendo Famicom / NES
+ Nintendo Game Boy (Color)
+ Nintendo Game Boy Advance
+ Nintendo Super Famicom / SNES
+ Nintendo Virtual Boy
+ SNK Neo Geo Pocket (Color)
+ Sega CD
+ Sega Game Gear
+ Sega Master System
+ Sega Mega Drive / Genesis
+ Sega Saturn
+ Sinclair ZX Spectrum
+ Sony PlayStation
+ Sony PlayStation Portable

## Quick installation with script

Prerequisites:

* A server with public IP address (for example, basic droplet on DO) running Ubuntu 22.04
* A domain with `A` record with IP address of the server

For example domain is `playtime.example.com` and IP is `10.10.10.10`.

Ensure `A` record is correct:

```bash
dig +short playtime.example.com
#should output correct IP address
```

On the server, execute:

```bash
mkdir playtime && cd playtime
wget "https://github.com/n-at/playtime/raw/master/docker/quick-install.sh"
chmod +x quick-install.sh
./quick-install.sh "10.10.10.10" "playtime.example.com"
```

This script will:

1. Install docker (if it is not installed)
2. Run [coturn](https://github.com/coturn/coturn), a TURN/STUN/ICE server
3. Get SSL certificate from [Let's Encrypt](https://letsencrypt.org/)
4. Build and run playtime

`admin` password will be in `playtime/data/admin.password` file.

## Building

Go 1.24+ and npm 7+ required.

```bash
./install.sh
go build -a -o app
```

## Configuration

Commandline arguments available:

```
$ ./app -help
Usage of ./app:
  -db-path string
        db path (default "data/bolt.db")
  -debug-emulator
        debug emulator (extended browser console output)
  -debug-netplay
        debug netplay (extended browser console output)
  -debug-templates
        debug page templates (do not cache)
  -listen string
        address and port to listen (default ":3000")
  -turn-server-password string
        TURN/STUN/ICE server password (if required)
  -turn-server-url string
        TURN/STUN/ICE server host, required for netplay (example: turn:turn.example.com)
  -turn-server-user string
        TURN/STUN/ICE server user name (if required)
  -uploads-path string
        local uploads path, ignored when -s3-bucket is set (default "uploads")
  -verbose
        show debug output
  -s3-bucket string
        S3 bucket name (enables S3 storage when set)
  -s3-region string
        S3 region (default "us-east-1")
  -s3-endpoint string
        S3 custom endpoint URL, for MinIO or S3-compatible services
  -s3-access-key-id string
        S3 access key ID
  -s3-secret-access-key string
        S3 secret access key
  -s3-use-path-style
        use path-style S3 URLs, required for MinIO
```

## S3 storage

By default playtime stores uploaded files in a local directory (`-uploads-path`). Setting `-s3-bucket` switches to S3-compatible object storage instead.

Game ROMs and save states are stored as objects under their ID-based path. Metadata (game info, save state info) is stored as JSON under `_meta/game/` and `_meta/savestate/` prefixes. On startup, the database is automatically rebuilt from these metadata files — so losing the local database is not a problem.

Example with AWS S3:

```bash
./app -s3-bucket my-playtime-bucket -s3-region eu-west-1 \
      -s3-access-key-id AKIAIOSFODNN7EXAMPLE \
      -s3-secret-access-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

Example with MinIO:

```bash
./app -s3-bucket playtime -s3-endpoint http://minio:9000 \
      -s3-access-key-id minioadmin -s3-secret-access-key minioadmin \
      -s3-use-path-style
```

If no credentials are provided, the AWS default credential chain is used (instance profile, `~/.aws/credentials`, environment variables, etc.).

## docker

Build an image:

```bash
docker image build -t playtime:latest .
```

Available image environment variables:

* `PLAYTIME_TURN_URL`
* `PLAYTIME_TURN_USER`
* `PLAYTIME_TURN_PASSWORD`
* `PLAYTIME_DEBUG_EMULATOR`
* `PLAYTIME_DEBUG_TEMPLATES`
* `PLAYTIME_DEBUG_NETPLAY`
* `PLAYTIME_VERBOSE`

Exposed volumes:

* `/app/data` - database directory
* `/app/uploads` - uploads directory (not needed when using S3)

Exposed default port `3000`.

Example `docker-compose.yml` with S3:

```yaml
services:
  playtime:
    image: playtime:latest
    volumes:
      - "./data:/data"
    ports:
      - 3000:3000
    command:
      - "--s3-bucket=my-playtime-bucket"
      - "--s3-region=eu-west-1"
      - "--s3-access-key-id=AKIAIOSFODNN7EXAMPLE"
      - "--s3-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

## Netplay

TURN server is required for netplay. It can be obtained from:

* [This list](https://gist.github.com/sagivo/3a4b2f2c7ac6e1b5267c2f1f59ac6c6b)
* [Open Relay](https://www.metered.ca/tools/openrelay/)
* Hosted, for example [coturn](https://github.com/coturn/coturn)

For a particular game netplay needs to be enabled in game settings.

## Uses

* [labstack/echo](https://github.com/labstack/echo) - MIT
* [flosch/pongo2](https://github.com/flosch/pongo2) - MIT
* [sirupsen/logrus](https://github.com/sirupsen/logrus) - MIT
* [timshannon/bolthold](https://github.com/timshannon/bolthold) - MIT
* [google/uuid](https://github.com/google/uuid) - BSD-3-Clause
* [twbs/bootstrap](https://github.com/twbs/bootstrap) - MIT
* [twbs/icons](https://github.com/twbs/icons) - MIT
* [sumimakito/Awesome-qr.js](https://github.com/sumimakito/Awesome-qr.js) - Apache-2.0
* [lekoala/bootstrap5-tags](https://github.com/lekoala/bootstrap5-tags) - MIT
* [@fontsource/comfortaa](https://www.npmjs.com/package/@fontsource/comfortaa) - OFL-1.1
* [@fontsource/raleway](https://www.npmjs.com/package/@fontsource/raleway) - OFL-1.1
* [EmulatorJS](https://github.com/EmulatorJS/EmulatorJS) - GPL-3.0 (not included)
* BIOS files from respective vendors (not included)
