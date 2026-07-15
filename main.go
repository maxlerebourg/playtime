package main

import (
	"embed"
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"playtime/storage"
	"playtime/web"
	"strings"
)

//go:embed templates
var staticTemplates embed.FS
//go:embed assets
var staticAssets embed.FS

var (
	storageConfig *storage.Configuration
	webConfig     *web.Configuration
)

func main() {
	s, err := storage.New(storageConfig)
	if err != nil {
		log.Fatalf("unable to open storage db: %s", err)
	}

	err = s.UserEnsureExists()
	if err != nil {
		log.Fatalf("unable to create default user: %s", err)
	}

	if err := s.SyncFromS3(); err != nil {
		log.Fatalf("unable to sync from S3: %s", err)
	}

	server := web.New(webConfig, s)
	log.Fatal(server.Start())
}

func init() {
	verbosePtr := flag.Bool("verbose", false, "show debug output")
	debugTemplatesPtr := flag.Bool("debug-templates", false, "debug page templates (do not cache)")
	debugEmulatorPtr := flag.Bool("debug-emulator", false, "debug emulator (extended browser console output)")
	debugNetplayPtr := flag.Bool("debug-netplay", false, "debug netplay (extended browser console output)")
	listenPtr := flag.String("listen", ":3000", "address and port to listen")
	dbPathPtr := flag.String("db-path", "data/bolt.db", "db path")
	uploadsPathPtr := flag.String("uploads-path", "uploads", "local uploads path (ignored when --s3-bucket is set)")
	turnServerUrlPtr := flag.String("turn-server-url", "", "TURN/STUN/ICE server host, required for netplay (example: turn:turn.example.com)")
	turnServerUser := flag.String("turn-server-user", "", "TURN/STUN/ICE server user name (if required)")
	turnServerPassword := flag.String("turn-server-password", "", "TURN/STUN/ICE server password (if required)")
	s3BucketPtr := flag.String("s3-bucket", "", "S3 bucket name (enables S3 storage when set)")
	s3RegionPtr := flag.String("s3-region", "us-east-1", "S3 region")
	s3EndpointPtr := flag.String("s3-endpoint", "", "S3 custom endpoint URL (for S3-compatible services)")
	s3AccessKeyIdPtr := flag.String("s3-access-key-id", "", "S3 access key ID")
	s3SecretAccessKeyPtr := flag.String("s3-secret-access-key", "", "S3 secret access key")
	s3UsePathStylePtr := flag.Bool("s3-use-path-style", false, "use path-style S3 URLs")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	if *verbosePtr {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *verbosePtr {
		log.Info("verbose output enabled")
	}
	if *debugTemplatesPtr {
		log.Info("templates debug enabled")
	}
	if *debugEmulatorPtr {
		log.Info("emulator debug enabled")
	}
	if *debugNetplayPtr {
		log.Info("netplay debug enabled")
	}

	storageConfig = &storage.Configuration{
		DatabasePath: *dbPathPtr,
		UploadsPath:  *uploadsPathPtr,
	}
	if strings.TrimSpace(*s3BucketPtr) != "" {
		storageConfig.S3 = &storage.S3Configuration{
			Bucket:          *s3BucketPtr,
			Region:          *s3RegionPtr,
			Endpoint:        *s3EndpointPtr,
			AccessKeyId:     *s3AccessKeyIdPtr,
			SecretAccessKey: *s3SecretAccessKeyPtr,
			UsePathStyle:    *s3UsePathStylePtr,
		}
	}
	webConfig = &web.Configuration{
		AssetsRoot: "assets",
		AssetsFS:   staticAssets,
		Listen:     *listenPtr,

		TemplatesDebug:     *debugTemplatesPtr,
		TemplatesRoot:      "templates",
		TemplatesFS:        staticTemplates,
		TemplatesExtension: "twig",

		EmulatorDebug: *debugEmulatorPtr,

		NetplayEnabled:     strings.TrimSpace(*turnServerUrlPtr) != "",
		NetplayDebug:       *debugNetplayPtr,
		TurnServerUrl:      *turnServerUrlPtr,
		TurnServerUser:     *turnServerUser,
		TurnServerPassword: *turnServerPassword,
	}
}
