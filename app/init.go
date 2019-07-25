package app

import (
	"encoding/json"
	"fmt"
	"sozluk/app/models"
	"sozluk/app/helpers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/revel/revel"
	"html/template"
	"strings"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

var DB *gorm.DB

func InitDB() {

	dbDriver := revel.Config.StringDefault("db.driver", "mysql")
	dbName := revel.Config.StringDefault("db.name", "sozluk")
	dbUser := revel.Config.StringDefault("db.user", "root")
	dbPassword := revel.Config.StringDefault("db.password", "root")
	dbAddress := revel.Config.StringDefault("db.address", "127.0.0.1")
	dbPort := revel.Config.StringDefault("db.port", "3306")

	var connstring string
	if dbDriver == "mysql" {
		connstring = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", dbUser, dbPassword, dbAddress, dbPort, dbName)
	} else if dbDriver == "postgres" {
		connstring = fmt.Sprintf("host=myhost port=myport user=gorm dbname=gorm password=mypassword")
	} else if dbDriver == "sqlite3" {
		connstring = dbAddress
	}

	var err error
	DB, err = gorm.Open(dbDriver, connstring)
	if err != nil {
		revel.AppLog.Info("DB Error", err)
	}
	revel.AppLog.Info("DB Connected")
	DB.AutoMigrate(&models.UserModel{})
	DB.AutoMigrate(&models.PostModel{})
	DB.AutoMigrate(&models.TopicModel{})
	DB.AutoMigrate(&models.ChannelModel{})
}

var DefaultLocale string
var GetDefaultLocaleFilter = func(c *revel.Controller, fc []revel.Filter) {
	DefaultLocale = c.Request.Locale
	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//for using outside of controller
func Trans(msg string, args ...interface{}) string {
	return revel.Message(DefaultLocale, msg, args)
}

var CurrentUser *models.UserModel
var GetCurrenUser = func(c *revel.Controller, fc []revel.Filter) {

	u := models.UserModel{}

	if err := json.Unmarshal([]byte(c.Session["user"].(string)), &u); err == nil {
		CurrentUser = &u
	}

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		GetDefaultLocaleFilter,        // Get Default Locale
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.BeforeAfterFilter,       // Call the before and after filter functions
		revel.ActionInvoker,           // Invoke the action.
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)

	//read config function for templates
	revel.TemplateFuncs["config"] = func(a string, b string) string {
		return revel.Config.StringDefault(a, b)
	}

	revel.TemplateFuncs["user"] = func() string {
		return CurrentUser.Username
	}

	revel.TemplateFuncs["format"] = func(str string) template.HTML {
		return template.HTML(strings.Replace(helpers.FormatContent(str), "\n", "<br>", -1))
	}

	revel.OnAppStart(InitDB)
}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")
	c.Response.Out.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}
