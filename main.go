package main

import (
	"better-console-backend/adapters"
	"better-console-backend/config"
	"better-console-backend/controllers"
	"better-console-backend/domain"
	"better-console-backend/domain/member"
	"better-console-backend/domain/organization"
	"better-console-backend/domain/rbac"
	"better-console-backend/domain/site"
	"better-console-backend/domain/webhook"
	"better-console-backend/middlewares"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/keepeye/logrus-filename"
	"github.com/labstack/echo"
	echomiddleware "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"time"
)

const (
	Version = "0.0.1"
	website = "https://www.bettercode.kr"
	banner  = `
    __         __  __                                           __   
   / /_  ___  / /_/ /____  _____   _________  ____  _________  / /__ 
  / __ \/ _ \/ __/ __/ _ \/ ___/  / ___/ __ \/ __ \/ ___/ __ \/ / _ \
 / /_/ /  __/ /_/ /_/  __/ /     / /__/ /_/ / / / (__  ) /_/ / /  __/
/_.___/\___/\__/\__/\___/_/      \___/\____/_/ /_/____/\____/_/\___/  
`
)

const (
	EnvBetterConsoleDbHost     = "BETTER_CONSOLE_DB_HOST"
	EnvBetterConsoleDbDriver   = "BETTER_CONSOLE_DB_DRIVER"
	EnvBetterConsoleDbName     = "BETTER_CONSOLE_DB_NAME"
	EnvBetterConsoleDbUser     = "BETTER_CONSOLE_DB_USER"
	EnvBetterConsoleDbPassword = "BETTER_CONSOLE_DB_PASSWORD"
)

var (
	gormDB   *gorm.DB
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
)

func init() {
	config.InitConfig("config/config.json")

	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	log.AddHook(filenameHook)
	log.SetFormatter(&log.JSONFormatter{DisableHTMLEscape: true})

	var dialector gorm.Dialector
	if len(os.Getenv(EnvBetterConsoleDbHost)) > 0 &&
		len(os.Getenv(EnvBetterConsoleDbDriver)) > 0 &&
		len(os.Getenv(EnvBetterConsoleDbName)) > 0 &&
		len(os.Getenv(EnvBetterConsoleDbUser)) > 0 &&
		len(os.Getenv(EnvBetterConsoleDbPassword)) > 0 {

		driver := os.Getenv(EnvBetterConsoleDbDriver)
		if driver != "mysql" {
			panic("not supported database")
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True",
			os.Getenv(EnvBetterConsoleDbUser),
			os.Getenv(EnvBetterConsoleDbPassword),
			os.Getenv(EnvBetterConsoleDbHost),
			os.Getenv(EnvBetterConsoleDbName))
		dialector = mysql.Open(dsn)
	} else {
		// 기본적으로 DB는 sqlite
		dialector = sqlite.Open("better-console.db")
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("Database Connection Error")
	}

	if err := initializeDatabase(db); err != nil {
		panic(fmt.Sprintf("Database initializeDatabase Error : %s", err.Error()))
	}

	sqlDB, err := db.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	gormDB = db
}

func initializeDatabase(db *gorm.DB) error {
	fmt.Println(">>> InitializeDatabase")
	// 테이블 생성
	if err := db.AutoMigrate(&member.MemberEntity{}, &site.SettingEntity{}, &rbac.PermissionEntity{},
		&rbac.RoleEntity{}, &organization.OrganizationEntity{},
		&webhook.WebHookEntity{}, &webhook.WebHookMessageEntity{}); err != nil {
		return err
	}

	var permissionCount int64
	db.Raw("SELECT count(*) FROM permissions WHERE type= 'pre-define'").Scan(&permissionCount)

	if permissionCount == 0 {
		if err := db.Exec("INSERT INTO permissions(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", domain.PermissionManageSystemSettings, "시스템 설정(예. 두레이 로그인 등) 권한", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO permissions(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", domain.PermissionManageMembers, "멤버 관리 권한", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO permissions(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", domain.PermissionManageAccessControl, "접근 제어 관리 권한", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO permissions(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", domain.PermissionManageOrganization, "조직 관리 권한", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO permissions(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", domain.PermissionNoteWebHooks, "웹훅 전송 권한", time.Now(), time.Now()).Error; err != nil {
			return err
		}
	}

	var roleCount int64
	db.Raw("SELECT count(*) FROM roles WHERE type= 'pre-define'").Scan(&roleCount)

	if roleCount == 0 {
		if err := db.Exec("INSERT INTO roles(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", "시스템 관리자", "", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO role_permissions(role_entity_id, permission_entity_id) values(1, 1)").Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO roles(type, name, description, created_at, updated_at) values(?, ?, ?, ?, ?)",
			"pre-define", "조직/멤버 관리자", "", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		if err := db.Exec("INSERT INTO role_permissions(role_entity_id, permission_entity_id) values(2, 2),(2, 3),(2, 4)").Error; err != nil {
			return err
		}
	}

	// siteadm 계정 만들기
	var signId string
	db.Raw("SELECT sign_id FROM members WHERE sign_id = ?", "siteadm").Scan(&signId)

	if len(signId) == 0 {
		if err := db.Exec("INSERT INTO members(type, sign_id, name, password, status, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?)",
			"site", "siteadm", "사이트 관리자", "$2a$04$7Ca1ybGc4yFkcBnzK1C0qevHy/LSD7PuBbPQTZEs6tiNM4hAxSYiG", "approved", time.Now(), time.Now()).Error; err != nil {
			return err
		}

		// 사이트 관리자에 사전 정의된 두가지 역할을 할당한다.(시스템 관리자, 멤버 관리자)
		if err := db.Exec("INSERT INTO member_roles(member_entity_id, role_entity_id) values(1, 1),(1, 2)").Error; err != nil {
			return err
		}
	}

	return nil
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Pre(echomiddleware.RemoveTrailingSlash())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowCredentials: true,
	}))

	e.Use(middlewares.GORMDb(gormDB))
	e.Use(middlewares.JwtToken())
	e.HideBanner = true

	e.GET("/ws/:id", connectWebSocket)
	controllers.AuthController{}.Init(e.Group("/api/auth"))
	controllers.SiteController{}.Init(e.Group("/api/site"))
	controllers.MemberController{}.Init(e.Group("/api/members"))
	controllers.AccessControlController{}.Init(e.Group("/api/access-control"))
	controllers.OrganizationController{}.Init(e.Group("/api/organizations"))
	controllers.WebHookController{}.Init(e.Group("/api/web-hooks"))
	controllers.WebHookGoController{}.Init(e.Group("/api/web-hook-go"))

	color.Println(banner, color.Red("v"+Version), color.Blue(website))
	e.Start(":2021")
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func connectWebSocket(ctx echo.Context) error {
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}

	webSocketId := ctx.Param("id")
	adapters.WebSocketAdapter().AddConnection(webSocketId, ws)

	return nil
}
