package retro

import (
	"context"
	"fmt"
	"strings"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/chewedfeed/soon-cms/internal/config"
	pgx "github.com/jackc/pgx/v4"
)

type CMS struct {
	Config       *config.Config
	CTX          context.Context
	ErrorChannel chan error
}

type Service struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	LaunchDate  LaunchDate `json:"launchDate"`
	Progress    int        `json:"progress"`
	Icon        string     `json:"icon"`
	FullDesc    string     `json:"fullDesc"`
	Uptime      string     `json:"uptime"`
}
type LaunchDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func NewCMS(config *config.Config, errChan chan error) *CMS {
	return &CMS{
		Config:       config,
		CTX:          context.Background(),
		ErrorChannel: errChan,
	}
}

func (c CMS) getDB() (*pgx.Conn, error) {
	conn, err := pgx.Connect(c.CTX, fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.Config.Database.User, c.Config.Database.Password, c.Config.Database.Host, c.Config.Database.Port, c.Config.Database.DBName))
	if err != nil {
		c.ErrorChannel <- bugLog.Error(err)
		return nil, bugLog.Error(err)
	}

	return conn, nil
}

func (c CMS) getServices() ([]Service, error) {
	db, err := c.getDB()
	if err != nil {
		c.ErrorChannel <- bugLog.Error(err)
		return nil, bugLog.Error(err)
	}
	defer db.Close(c.CTX)

	rows, err := db.Query(c.CTX, "SELECT name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime FROM services WHERE started = true")
	if err != nil {
		return nil, bugLog.Error(err)
	}
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var service Service
		if err := rows.Scan(
			&service.Name,
			&service.Description,
			&service.LaunchDate.Year,
			&service.LaunchDate.Month,
			&service.LaunchDate.Day,
			&service.URL,
			&service.Progress,
			&service.Icon,
			&service.FullDesc,
			&service.Uptime); err != nil {
			return nil, bugLog.Error(err)
		}
		service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)
		services = append(services, service)
	}

	return services, nil
}

func (c CMS) getService(name string) (Service, error) {
	db, err := c.getDB()
	if err != nil {
		c.ErrorChannel <- bugLog.Error(err)
		return Service{}, bugLog.Error(err)
	}
	defer db.Close(c.CTX)

	var service Service
	if err := db.QueryRow(c.CTX, "SELECT name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime FROM services WHERE search_name = $1", name).Scan(
		&service.Name,
		&service.Description,
		&service.LaunchDate.Year,
		&service.LaunchDate.Month,
		&service.LaunchDate.Day,
		&service.URL,
		&service.Progress,
		&service.Icon,
		&service.FullDesc,
		&service.Uptime); err != nil {
		return Service{}, bugLog.Error(err)
	}

	service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)
	return service, nil
}

func (c CMS) AllowedOrigins() ([]string, error) {
	db, err := c.getDB()
	if err != nil {
		c.ErrorChannel <- bugLog.Error(err)
		return nil, bugLog.Error(err)
	}
	defer db.Close(c.CTX)

	rows, err := db.Query(c.CTX, "SELECT url, alternatives FROM services")
	if err != nil {
		return nil, bugLog.Error(err)
	}
	type service struct {
		URL  *string
		Alts *string
	}

	defer rows.Close()
	origins := make([]string, 0)
	for rows.Next() {
		var s service
		if err := rows.Scan(&s.URL, &s.Alts); err != nil {
			return nil, bugLog.Error(err)
		}
		if s.Alts != nil {
			alts := strings.Split(*s.Alts, ",")
			for _, alt := range alts {
				origins = append(origins, fmt.Sprintf("https://%s", alt))
				origins = append(origins, fmt.Sprintf("https://www.%s", alt))
			}
		}
		origins = append(origins, *s.URL)

		www := strings.Replace(*s.URL, "https://", "https://www.", 1)
		origins = append(origins, www)
	}

	return origins, nil
}
