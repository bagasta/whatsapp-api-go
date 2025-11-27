package cmd

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainAgent "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	domainChat "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chat"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	domainGroup "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/group"
	domainMessage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/message"
	domainNewsletter "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/newsletter"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	domainSession "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
	domainUser "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/user"
	domainWebhook "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	EmbedIndex embed.FS
	EmbedViews embed.FS

	// Whatsapp
	whatsappCli   *whatsmeow.Client
	clientManager *whatsapp.ClientManager

	// Chat Storage
	chatStorageDB   *sql.DB
	chatStorageRepo domainChatStorage.IChatStorageRepository

	// Repositories
	sessionRepo repository.SessionRepository
	apiKeyRepo  repository.ApiKeyRepository
	webhookRepo repository.WebhookConfigRepository

	// Usecase
	appUsecase        domainApp.IAppUsecase
	chatUsecase       domainChat.IChatUsecase
	sendUsecase       domainSend.ISendUsecase
	userUsecase       domainUser.IUserUsecase
	messageUsecase    domainMessage.IMessageUsecase
	groupUsecase      domainGroup.IGroupUsecase
	newsletterUsecase domainNewsletter.INewsletterUsecase
	sessionUsecase    domainSession.ISessionUsecase
	agentUsecase      domainAgent.IAgentUsecase
	webhookUsecase    domainWebhook.IWebhookConfigUsecase
)

// reconnectExistingSessions initializes clients for all stored sessions and attempts to connect them.
func reconnectExistingSessions(ctx context.Context, cm *whatsapp.ClientManager, repo domainSession.ISessionRepository) {
	sessions, err := repo.List()
	if err != nil {
		logrus.Warnf("failed to list sessions for auto-reconnect: %v", err)
		return
	}

	for _, s := range sessions {
		agentID := s.AgentID
		go func(agent string) {
			client, err := cm.CreateClient(ctx, agent)
			if err != nil {
				logrus.Warnf("auto-reconnect: create client failed for agent %s: %v", agent, err)
				return
			}
			if client.IsConnected() {
				return
			}
			if err := client.Connect(); err != nil {
				logrus.Warnf("auto-reconnect: connect failed for agent %s: %v", agent, err)
			} else {
				logrus.Infof("auto-reconnect: agent %s connected", agent)
			}
		}(agentID)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Send free whatsapp API",
	Long: `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice, 
you can send whatsapp over http api but your whatsapp account have to be multi device version`,
}

func init() {
	// Load environment variables first
	utils.LoadConfig(".")

	time.Local = time.UTC

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize flags first, before any subcommands are added
	initFlags()

	// Then initialize other components
	cobra.OnInitialize(initEnvConfig, initApp)
}

// initEnvConfig loads configuration from environment variables
func initEnvConfig() {
	fmt.Println(viper.AllSettings())
	// Application settings
	if envPort := viper.GetString("app_port"); envPort != "" {
		config.AppPort = envPort
	}
	if envDebug := viper.GetBool("app_debug"); envDebug {
		config.AppDebug = envDebug
	}
	if envOs := viper.GetString("app_os"); envOs != "" {
		config.AppOs = envOs
	}
	if envBasicAuth := viper.GetString("app_basic_auth"); envBasicAuth != "" {
		credential := strings.Split(envBasicAuth, ",")
		config.AppBasicAuthCredential = credential
	}
	if envBasePath := viper.GetString("app_base_path"); envBasePath != "" {
		config.AppBasePath = envBasePath
	}
	if envTrustedProxies := viper.GetString("app_trusted_proxies"); envTrustedProxies != "" {
		proxies := strings.Split(envTrustedProxies, ",")
		config.AppTrustedProxies = proxies
	}

	// Chat storage settings
	if envChatStorageURI := viper.GetString("chat_storage_uri"); envChatStorageURI != "" {
		config.ChatStorageURI = envChatStorageURI
	}
	if viper.IsSet("chat_storage_enable_foreign_keys") {
		config.ChatStorageEnableForeignKeys = viper.GetBool("chat_storage_enable_foreign_keys")
	}
	if viper.IsSet("chat_storage_enable_wal") {
		config.ChatStorageEnableWAL = viper.GetBool("chat_storage_enable_wal")
	}

	// Database settings
	if envDBURI := viper.GetString("db_uri"); envDBURI != "" {
		config.DBURI = envDBURI
	}
	if envDBKEYSURI := viper.GetString("db_keys_uri"); envDBKEYSURI != "" {
		config.DBKeysURI = envDBKEYSURI
	}

	// WhatsApp settings
	if envAutoReply := viper.GetString("whatsapp_auto_reply"); envAutoReply != "" {
		config.WhatsappAutoReplyMessage = envAutoReply
	}
	if viper.IsSet("whatsapp_auto_mark_read") {
		config.WhatsappAutoMarkRead = viper.GetBool("whatsapp_auto_mark_read")
	}
	if viper.IsSet("whatsapp_auto_download_media") {
		config.WhatsappAutoDownloadMedia = viper.GetBool("whatsapp_auto_download_media")
	}
	if envWebhook := viper.GetString("whatsapp_webhook"); envWebhook != "" {
		webhook := strings.Split(envWebhook, ",")
		config.WhatsappWebhook = webhook
	}
	if envWebhookSecret := viper.GetString("whatsapp_webhook_secret"); envWebhookSecret != "" {
		config.WhatsappWebhookSecret = envWebhookSecret
	}
	if viper.IsSet("whatsapp_account_validation") {
		config.WhatsappAccountValidation = viper.GetBool("whatsapp_account_validation")
	}

	if envAiBackend := viper.GetString("ai_backend_url"); envAiBackend != "" {
		config.AiBackendURL = envAiBackend
	}
}

func initFlags() {
	// Application flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AppPort,
		"port", "p",
		config.AppPort,
		"change port number with --port <number> | example: --port=8080",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&config.AppDebug,
		"debug", "d",
		config.AppDebug,
		"hide or displaying log with --debug <true/false> | example: --debug=true",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppOs,
		"os", "",
		config.AppOs,
		`os name --os <string> | example: --os="Chrome"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppBasicAuthCredential,
		"basic-auth", "b",
		config.AppBasicAuthCredential,
		"basic auth credential | -b=yourUsername:yourPassword",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppBasePath,
		"base-path", "",
		config.AppBasePath,
		`base path for subpath deployment --base-path <string> | example: --base-path="/gowa"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppTrustedProxies,
		"trusted-proxies", "",
		config.AppTrustedProxies,
		`trusted proxy IP ranges for reverse proxy deployments --trusted-proxies <string> | example: --trusted-proxies="0.0.0.0/0" or --trusted-proxies="10.0.0.0/8,172.16.0.0/12"`,
	)

	// Chat storage flags
	rootCmd.PersistentFlags().StringVar(
		&config.ChatStorageURI,
		"chat-storage-uri",
		config.ChatStorageURI,
		`database uri for chat/session storage (default sqlite at storages/chatstorage.db). example: --chat-storage-uri="postgres://user:password@localhost:5432/whatsapp_db?sslmode=disable"`,
	)
	rootCmd.PersistentFlags().BoolVar(
		&config.ChatStorageEnableForeignKeys,
		"chat-storage-enable-foreign-keys",
		config.ChatStorageEnableForeignKeys,
		`enable SQLite foreign keys for chat/session storage --chat-storage-enable-foreign-keys=<true/false>`,
	)
	rootCmd.PersistentFlags().BoolVar(
		&config.ChatStorageEnableWAL,
		"chat-storage-enable-wal",
		config.ChatStorageEnableWAL,
		`enable SQLite WAL mode for chat/session storage --chat-storage-enable-wal=<true/false>`,
	)

	// Database flags
	rootCmd.PersistentFlags().StringVarP(
		&config.DBURI,
		"db-uri", "",
		config.DBURI,
		`the database uri to store the connection data database uri (by default, we'll use sqlite3 under storages/whatsapp.db). database uri --db-uri <string> | example: --db-uri="file:storages/whatsapp.db?_foreign_keys=on or postgres://user:password@localhost:5432/whatsapp"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.DBKeysURI,
		"db-keys-uri", "",
		config.DBKeysURI,
		`the database uri to store the keys database uri (by default, we'll use the same database uri). database uri --db-keys-uri <string> | example: --db-keys-uri="file::memory:?cache=shared&_foreign_keys=on"`,
	)

	// WhatsApp flags
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappAutoReplyMessage,
		"autoreply", "",
		config.WhatsappAutoReplyMessage,
		`auto reply when received message --autoreply <string> | example: --autoreply="Don't reply this message"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoMarkRead,
		"auto-mark-read", "",
		config.WhatsappAutoMarkRead,
		`auto mark incoming messages as read --auto-mark-read <true/false> | example: --auto-mark-read=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoDownloadMedia,
		"auto-download-media", "",
		config.WhatsappAutoDownloadMedia,
		`auto download media from incoming messages --auto-download-media <true/false> | example: --auto-download-media=false`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.WhatsappWebhook,
		"webhook", "w",
		config.WhatsappWebhook,
		`forward event to webhook --webhook <string> | example: --webhook="https://yourcallback.com/callback"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappWebhookSecret,
		"webhook-secret", "",
		config.WhatsappWebhookSecret,
		`secure webhook request --webhook-secret <string> | example: --webhook-secret="super-secret-key"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAccountValidation,
		"account-validation", "",
		config.WhatsappAccountValidation,
		`enable or disable account validation --account-validation <true/false> | example: --account-validation=true`,
	)
}

func initChatStorage() (*sql.DB, error) {
	var driverName string
	var connStr string

	if strings.HasPrefix(config.ChatStorageURI, "postgres://") || strings.HasPrefix(config.ChatStorageURI, "postgresql://") {
		driverName = "postgres"
		connStr = config.ChatStorageURI
	} else {
		driverName = "sqlite3"
		connStr = fmt.Sprintf("%s?_journal_mode=WAL", config.ChatStorageURI)
		if config.ChatStorageEnableForeignKeys {
			connStr += "&_foreign_keys=on"
		}
	}

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := database.Migrate(db, driverName); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func initApp() {
	if config.AppDebug {
		config.WhatsappLogLevel = "DEBUG"
		logrus.SetLevel(logrus.DebugLevel)
	}

	//preparing folder if not exist
	err := utils.CreateFolder(config.PathQrCode, config.PathSendItems, config.PathStorages, config.PathMedia)
	if err != nil {
		logrus.Errorln(err)
	}

	ctx := context.Background()

	chatStorageDB, err = initChatStorage()
	if err != nil {
		// Terminate the application if chat storage fails to initialize to avoid nil pointer panics later.
		logrus.Fatalf("failed to initialize chat storage: %v", err)
	}

	isChatPostgres := strings.HasPrefix(config.ChatStorageURI, "postgres://") || strings.HasPrefix(config.ChatStorageURI, "postgresql://")
	chatStorageRepo = chatstorage.NewStorageRepository(chatStorageDB, isChatPostgres)
	if err := chatStorageRepo.InitializeSchema(); err != nil {
		logrus.Fatalf("failed to initialize chat storage schema: %v", err)
	}

	webhookRepo = *repository.NewWebhookConfigRepository(chatStorageDB).(*repository.WebhookConfigRepository)
	webhookUsecase = domainWebhook.NewWebhookConfigUsecase(&webhookRepo, &sessionRepo)
	if err := webhookUsecase.SyncRuntimeConfig(); err != nil {
		logrus.Warnf("failed to sync webhook config from DB: %v", err)
	}
	whatsapp.SetWebhookResolver(webhookUsecase.ResolveWebhooks)

	whatsappDB := whatsapp.InitWaDB(ctx, config.DBURI)
	var keysDB *sqlstore.Container
	if config.DBKeysURI != "" {
		keysDB = whatsapp.InitWaDB(ctx, config.DBKeysURI)
	}

	whatsappCli = whatsapp.InitWaCLI(ctx, whatsappDB, keysDB, chatStorageRepo)

	// Initialize ClientManager for multi-agent support
	clientManager = whatsapp.NewClientManager(chatStorageRepo)

	// Initialize repositories
	sessionRepo = *repository.NewSessionRepository(chatStorageDB).(*repository.SessionRepository)
	apiKeyRepo = *repository.NewApiKeyRepository(chatStorageDB).(*repository.ApiKeyRepository)

	// Usecase
	appUsecase = usecase.NewAppService(chatStorageRepo)
	chatUsecase = usecase.NewChatService(chatStorageRepo)
	sendUsecase = usecase.NewSendService(appUsecase, chatStorageRepo)
	userUsecase = usecase.NewUserService()
	messageUsecase = usecase.NewMessageService(chatStorageRepo)
	groupUsecase = usecase.NewGroupService()
	newsletterUsecase = usecase.NewNewsletterService()
	sessionUsecase = domainSession.NewSessionUsecase(&sessionRepo, &apiKeyRepo, clientManager)
	agentUsecase = domainAgent.NewAgentUsecase(&sessionRepo, &apiKeyRepo, clientManager)

	// Auto-reconnect previously saved sessions
	reconnectExistingSessions(ctx, clientManager, &sessionRepo)

	// Auto-forward inbound messages to AI for multi-agent clients
	whatsapp.SetAgentForwarder(func(agentID string, evt *events.Message) {
		// Skip self or broadcast messages
		chatJID := evt.Info.Chat.String()
		if evt.Info.IsFromMe || evt.Info.IsIncomingBroadcast() {
			return
		}

		// Resolve client & self JID for mention detection
		client := clientManager.GetClient(agentID)
		selfJID := ""
		selfAltJID := ""
		if client != nil && client.Store != nil && client.Store.ID != nil {
			selfJID = client.Store.ID.String()
			if alt, err := client.Store.GetAltJID(ctx, *client.Store.ID); err == nil && alt.String() != "" {
				selfAltJID = alt.String()
			}
		}

		// Extract human text
		extractText := func(msg *waE2E.Message) string {
			inner := msg
			for i := 0; i < 3; i++ {
				switch {
				case inner.GetViewOnceMessage() != nil && inner.GetViewOnceMessage().GetMessage() != nil:
					inner = inner.GetViewOnceMessage().GetMessage()
					continue
				case inner.GetEphemeralMessage() != nil && inner.GetEphemeralMessage().GetMessage() != nil:
					inner = inner.GetEphemeralMessage().GetMessage()
					continue
				case inner.GetViewOnceMessageV2() != nil && inner.GetViewOnceMessageV2().GetMessage() != nil:
					inner = inner.GetViewOnceMessageV2().GetMessage()
					continue
				case inner.GetViewOnceMessageV2Extension() != nil && inner.GetViewOnceMessageV2Extension().GetMessage() != nil:
					inner = inner.GetViewOnceMessageV2Extension().GetMessage()
					continue
				}
				break
			}

			if conv := inner.GetConversation(); conv != "" {
				return conv
			}
			if ext := inner.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
				return ext.GetText()
			}
			if protoMsg := inner.GetProtocolMessage(); protoMsg != nil {
				if edited := protoMsg.GetEditedMessage(); edited != nil {
					if ext := edited.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
						return ext.GetText()
					}
					if conv := edited.GetConversation(); conv != "" {
						return conv
					}
				}
			}
			return ""
		}

		text := extractText(evt.Message)
		if text == "" {
			return
		}

		// For group chats, only forward if:
		// - bot is mentioned (ContextInfo),
		// - text contains bare self user / phone,
		// - user is replying to bot,
		// - OR (fallback) there is any mention in the message (so mentions that differ in JID format still pass).
		if utils.IsGroupJID(chatJID) {
			selfUser := ""
			selfPhone := ""
			if j, err := types.ParseJID(selfJID); err == nil {
				selfUser = j.User
				selfPhone = strings.TrimPrefix(j.User, "+")
			}
			mentioned := isMentioned(evt.Message, selfJID) || isMentioned(evt.Message, selfAltJID)
			containsSelf := selfUser != "" && strings.Contains(text, selfUser)
			containsPhone := selfPhone != "" && strings.Contains(text, selfPhone)
			repliesToBot := isReplyToSelf(evt.Message, selfJID)
			if !mentioned && !containsSelf && !containsPhone && !repliesToBot {
				logrus.Debugf("Auto-forward AI skipped (group, no mention): agent=%s chat=%s self=%s text=%q", agentID, chatJID, selfUser, text)
				return
			}
		}

		// Resolve API key from session record
		sessionData, err := sessionRepo.FindByAgentID(agentID)
		if err != nil {
			logrus.Warnf("Auto-forward AI: session lookup failed for agent %s: %v", agentID, err)
			return
		}
		if sessionData == nil || sessionData.ApiKey == "" {
			logrus.Warnf("Auto-forward AI: no API key for agent %s, skipping", agentID)
			return
		}

		_, err = agentUsecase.ExecuteRun(agentID, sessionData.ApiKey, domainAgent.RunRequest{
			Input:      text,
			SessionID:  chatJID,
			Parameters: map[string]any{"max_steps": 3},
		})
		if err != nil {
			logrus.Warnf("Auto-forward AI: ExecuteRun failed for agent %s: %v", agentID, err)
		}
	})
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(embedIndex embed.FS, embedViews embed.FS) {
	EmbedIndex = embedIndex
	EmbedViews = embedViews
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// unwrapMessage walks through wrappers (view once / ephemeral) to reach inner content.
func unwrapMessage(msg *waE2E.Message) *waE2E.Message {
	if msg == nil {
		return nil
	}
	inner := msg
	for i := 0; i < 3; i++ {
		switch {
		case inner.GetViewOnceMessage() != nil && inner.GetViewOnceMessage().GetMessage() != nil:
			inner = inner.GetViewOnceMessage().GetMessage()
			continue
		case inner.GetEphemeralMessage() != nil && inner.GetEphemeralMessage().GetMessage() != nil:
			inner = inner.GetEphemeralMessage().GetMessage()
			continue
		case inner.GetViewOnceMessageV2() != nil && inner.GetViewOnceMessageV2().GetMessage() != nil:
			inner = inner.GetViewOnceMessageV2().GetMessage()
			continue
		case inner.GetViewOnceMessageV2Extension() != nil && inner.GetViewOnceMessageV2Extension().GetMessage() != nil:
			inner = inner.GetViewOnceMessageV2Extension().GetMessage()
			continue
		}
		break
	}
	return inner
}

// isMentioned returns true if message mentions targetJID (case-insensitive).
func isMentioned(msg *waE2E.Message, targetJID string) bool {
	if targetJID == "" || msg == nil {
		return false
	}
	inner := unwrapMessage(msg)
	if inner == nil {
		return false
	}
	if ext := inner.GetExtendedTextMessage(); ext != nil {
		if ci := ext.GetContextInfo(); ci != nil {
			targetDigits := phoneDigits(targetJID)
			for _, jid := range ci.GetMentionedJID() {
				if equalBareJID(jid, targetJID) {
					return true
				}
				if targetDigits != "" && phoneDigits(jid) == targetDigits {
					return true
				}
			}
		}
	}
	return false
}

// equalBareJID compares two JIDs ignoring device/resource part.
func equalBareJID(a, b string) bool {
	parse := func(s string) (types.JID, bool) {
		j, err := types.ParseJID(s)
		if err != nil {
			return types.JID{}, false
		}
		return j, true
	}
	ja, oka := parse(a)
	jb, okb := parse(b)
	if oka && okb {
		return strings.EqualFold(ja.User, jb.User) && strings.EqualFold(ja.Server, jb.Server)
	}
	// Fallback: strip device part "user:device@server"
	strip := func(s string) string {
		if idx := strings.Index(s, ":"); idx != -1 {
			if at := strings.Index(s, "@"); at != -1 {
				return s[:idx] + s[at:]
			}
		}
		return s
	}
	return strings.EqualFold(strip(a), strip(b))
}

func phoneDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// isReplyToSelf checks if the message quotes a message from selfJID.
func isReplyToSelf(msg *waE2E.Message, selfJID string) bool {
	if selfJID == "" || msg == nil {
		return false
	}
	inner := unwrapMessage(msg)
	if inner == nil {
		return false
	}
	if ext := inner.GetExtendedTextMessage(); ext != nil {
		if ci := ext.GetContextInfo(); ci != nil {
			if ci.GetParticipant() != "" && equalBareJID(ci.GetParticipant(), selfJID) {
				return true
			}
		}
	}
	return false
}
