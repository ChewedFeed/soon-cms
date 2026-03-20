package retro

import (
	"context"
	"fmt"
	"strings"

	"github.com/bugfixes/go-bugfixes/logs"
	flagsService "github.com/flags-gg/go-flags"
	ConfigBuilder "github.com/keloran/go-config"
)

type CMS struct {
	Config       *ConfigBuilder.Config
	Flags        *flagsService.Client
	CTX          context.Context
	ErrorChannel chan error
}

type Service struct {
	ID          int            `json:"-"`
	Name        string         `json:"name"`
	SearchName  string         `json:"searchName"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	LaunchDate  LaunchDate     `json:"launchDate"`
	Progress    int            `json:"progress"`
	Progress2   float32        `json:"progress2"`
	Icon        string         `json:"icon"`
	FullDesc    string         `json:"fullDesc"`
	Uptime      string         `json:"uptime"`
	Launched    bool           `json:"launched"`
	Links       []ProjectLink  `json:"links,omitempty"`
	Roadmap     []RoadmapItem  `json:"roadmap,omitempty"`
}

type LaunchDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type ProjectLink struct {
	ID       int    `json:"id,omitempty"`
	LinkType string `json:"type"`
	URL      string `json:"url"`
	Label    string `json:"label,omitempty"`
}

type RoadmapItem struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name"`
	TargetDate  *string `json:"targetDate,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Completed   bool    `json:"completed"`
	SortOrder   int     `json:"sortOrder"`
}

type LaunchTask struct {
	ID        int  `json:"id,omitempty"`
	ServiceID int  `json:"serviceId"`
	Completed bool `json:"completed"`
}

type CreateServiceRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	FullDesc    string     `json:"fullDesc"`
	URL         string     `json:"url"`
	Icon        string     `json:"icon"`
	Uptime      string     `json:"uptime"`
	LaunchDate  LaunchDate `json:"launchDate"`
}

type CreateLinkRequest struct {
	LinkType string `json:"type"`
	URL      string `json:"url"`
	Label    string `json:"label"`
}

type CreateRoadmapRequest struct {
	Name       string  `json:"name"`
	TargetDate *string `json:"targetDate"`
	Completed  bool    `json:"completed"`
	SortOrder  int     `json:"sortOrder"`
}

type UpdateRoadmapRequest struct {
	Name        *string `json:"name,omitempty"`
	TargetDate  *string `json:"targetDate,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
	SortOrder   *int    `json:"sortOrder,omitempty"`
}

func NewCMS(config *ConfigBuilder.Config, flags *flagsService.Client, errChan chan error) *CMS {
	return &CMS{
		Config:       config,
		Flags:        flags,
		CTX:          context.Background(),
		ErrorChannel: errChan,
	}
}

func (c CMS) getServices() ([]Service, error) {
	services := make([]Service, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return services, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT id, name, search_name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime, live FROM services WHERE started = true")
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var service Service
		if err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.SearchName,
			&service.Description,
			&service.LaunchDate.Year,
			&service.LaunchDate.Month,
			&service.LaunchDate.Day,
			&service.URL,
			&service.Progress,
			&service.Icon,
			&service.FullDesc,
			&service.Uptime,
			&service.Launched); err != nil {
			return nil, logs.Error(err)
		}
		service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)
		services = append(services, service)
	}

	for i, svc := range services {
		links, err := c.getServiceLinks(svc.ID)
		if err != nil {
			return services, logs.Error(err)
		}
		services[i].Links = links
	}

	return services, nil
}

func (c CMS) getService(name string) (Service, error) {
	var service Service

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return service, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	if err := db.QueryRow(c.CTX, "SELECT id, name, search_name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime FROM services WHERE search_name = $1", name).Scan(
		&service.ID,
		&service.Name,
		&service.SearchName,
		&service.Description,
		&service.LaunchDate.Year,
		&service.LaunchDate.Month,
		&service.LaunchDate.Day,
		&service.URL,
		&service.Progress,
		&service.Icon,
		&service.FullDesc,
		&service.Uptime); err != nil {
		return Service{}, logs.Error(err)
	}

	service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)

	prog, err := c.getServiceProgress(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Progress2 = prog

	links, err := c.getServiceLinks(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Links = links

	roadmap, err := c.getServiceRoadmap(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Roadmap = roadmap

	return service, nil
}

func (c CMS) getServiceProgress(id int) (float32, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return 0, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()
	rows, err := db.Query(c.CTX, "SELECT completed FROM launch_task WHERE service_id = $1", id)
	if err != nil {
		return 0, logs.Error(err)
	}
	totalRows := 0
	completedRows := 0

	defer rows.Close()
	for rows.Next() {
		var completed bool
		if err := rows.Scan(&completed); err != nil {
			return 0, logs.Error(err)
		}
		totalRows++
		if completed {
			completedRows++
		}
	}
	if totalRows == 0 {
		return 0, nil
	}

	return (float32(completedRows) / float32(totalRows)) * 100, nil
}

func (c CMS) getServiceLinks(serviceID int) ([]ProjectLink, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT id, link_type, url, COALESCE(label, '') FROM project_links WHERE service_id = $1", serviceID)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var links []ProjectLink
	for rows.Next() {
		var link ProjectLink
		if err := rows.Scan(&link.ID, &link.LinkType, &link.URL, &link.Label); err != nil {
			return nil, logs.Error(err)
		}
		links = append(links, link)
	}

	return links, nil
}

func (c CMS) getServiceRoadmap(serviceID int) ([]RoadmapItem, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT id, name, target_date::text, release_date::text, completed, sort_order FROM project_roadmap WHERE service_id = $1 ORDER BY sort_order", serviceID)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var roadmap []RoadmapItem
	for rows.Next() {
		var item RoadmapItem
		if err := rows.Scan(&item.ID, &item.Name, &item.TargetDate, &item.ReleaseDate, &item.Completed, &item.SortOrder); err != nil {
			return nil, logs.Error(err)
		}
		roadmap = append(roadmap, item)
	}

	return roadmap, nil
}

// Write operations

func (c CMS) createService(req CreateServiceRequest) (Service, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return Service{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	searchName := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	var svc Service
	err = db.QueryRow(c.CTX,
		`INSERT INTO services (name, search_name, description, full_description, url, icon, uptime, launch_year, launch_month, launch_day, started)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
		 RETURNING id, name, description, url, launch_year, launch_month, launch_day, progress, icon, full_description, uptime, live`,
		req.Name, searchName, req.Description, req.FullDesc, req.URL, req.Icon, req.Uptime,
		req.LaunchDate.Year, req.LaunchDate.Month, req.LaunchDate.Day,
	).Scan(&svc.ID, &svc.Name, &svc.Description, &svc.URL,
		&svc.LaunchDate.Year, &svc.LaunchDate.Month, &svc.LaunchDate.Day,
		&svc.Progress, &svc.Icon, &svc.FullDesc, &svc.Uptime, &svc.Launched)
	if err != nil {
		return Service{}, logs.Error(err)
	}
	return svc, nil
}

func (c CMS) updateService(name string, req CreateServiceRequest) (Service, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return Service{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var svc Service
	err = db.QueryRow(c.CTX,
		`UPDATE services SET name=$1, description=$2, full_description=$3, url=$4, icon=$5, uptime=$6,
		 launch_year=$7, launch_month=$8, launch_day=$9
		 WHERE search_name=$10
		 RETURNING id, name, description, url, launch_year, launch_month, launch_day, progress, icon, full_description, uptime, live`,
		req.Name, req.Description, req.FullDesc, req.URL, req.Icon, req.Uptime,
		req.LaunchDate.Year, req.LaunchDate.Month, req.LaunchDate.Day, name,
	).Scan(&svc.ID, &svc.Name, &svc.Description, &svc.URL,
		&svc.LaunchDate.Year, &svc.LaunchDate.Month, &svc.LaunchDate.Day,
		&svc.Progress, &svc.Icon, &svc.FullDesc, &svc.Uptime, &svc.Launched)
	if err != nil {
		return Service{}, logs.Error(err)
	}
	return svc, nil
}

func (c CMS) deleteService(name string) error {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	_, err = db.Exec(c.CTX, "DELETE FROM services WHERE search_name = $1", name)
	if err != nil {
		return logs.Error(err)
	}
	return nil
}

func (c CMS) createLink(serviceName string, req CreateLinkRequest) (ProjectLink, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return ProjectLink{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var link ProjectLink
	err = db.QueryRow(c.CTX,
		`INSERT INTO project_links (service_id, link_type, url, label)
		 SELECT id, $2, $3, $4 FROM services WHERE search_name = $1
		 RETURNING id, link_type, url, COALESCE(label, '')`,
		serviceName, req.LinkType, req.URL, req.Label,
	).Scan(&link.ID, &link.LinkType, &link.URL, &link.Label)
	if err != nil {
		return ProjectLink{}, logs.Error(err)
	}
	return link, nil
}

func (c CMS) deleteLink(linkID int) error {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	_, err = db.Exec(c.CTX, "DELETE FROM project_links WHERE id = $1", linkID)
	if err != nil {
		return logs.Error(err)
	}
	return nil
}

func (c CMS) createRoadmapItem(serviceName string, req CreateRoadmapRequest) (RoadmapItem, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return RoadmapItem{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var item RoadmapItem
	err = db.QueryRow(c.CTX,
		`INSERT INTO project_roadmap (service_id, name, target_date, completed, sort_order)
		 SELECT id, $2, $3::date, $4, $5 FROM services WHERE search_name = $1
		 RETURNING id, name, target_date::text, release_date::text, completed, sort_order`,
		serviceName, req.Name, req.TargetDate, req.Completed, req.SortOrder,
	).Scan(&item.ID, &item.Name, &item.TargetDate, &item.ReleaseDate, &item.Completed, &item.SortOrder)
	if err != nil {
		return RoadmapItem{}, logs.Error(err)
	}
	return item, nil
}

func (c CMS) updateRoadmapItem(itemID int, req UpdateRoadmapRequest) (RoadmapItem, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return RoadmapItem{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var item RoadmapItem
	err = db.QueryRow(c.CTX,
		`UPDATE project_roadmap SET
		 name = COALESCE($2, name),
		 target_date = COALESCE($3::date, target_date),
		 release_date = COALESCE($4::date, release_date),
		 completed = COALESCE($5, completed),
		 sort_order = COALESCE($6, sort_order)
		 WHERE id = $1
		 RETURNING id, name, target_date::text, release_date::text, completed, sort_order`,
		itemID, req.Name, req.TargetDate, req.ReleaseDate, req.Completed, req.SortOrder,
	).Scan(&item.ID, &item.Name, &item.TargetDate, &item.ReleaseDate, &item.Completed, &item.SortOrder)
	if err != nil {
		return RoadmapItem{}, logs.Error(err)
	}
	return item, nil
}

func (c CMS) deleteRoadmapItem(itemID int) error {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	_, err = db.Exec(c.CTX, "DELETE FROM project_roadmap WHERE id = $1", itemID)
	if err != nil {
		return logs.Error(err)
	}
	return nil
}

func (c CMS) getLaunchTasks(serviceName string) ([]LaunchTask, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX,
		`SELECT lt.id, lt.service_id, lt.completed FROM launch_task lt
		 JOIN services s ON s.id = lt.service_id
		 WHERE s.search_name = $1`, serviceName)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var tasks []LaunchTask
	for rows.Next() {
		var t LaunchTask
		if err := rows.Scan(&t.ID, &t.ServiceID, &t.Completed); err != nil {
			return nil, logs.Error(err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (c CMS) createLaunchTask(serviceName string, completed bool) (LaunchTask, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return LaunchTask{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var t LaunchTask
	err = db.QueryRow(c.CTX,
		`INSERT INTO launch_task (service_id, completed)
		 SELECT id, $2 FROM services WHERE search_name = $1
		 RETURNING id, service_id, completed`,
		serviceName, completed,
	).Scan(&t.ID, &t.ServiceID, &t.Completed)
	if err != nil {
		return LaunchTask{}, logs.Error(err)
	}
	return t, nil
}

func (c CMS) updateLaunchTask(taskID int, completed bool) (LaunchTask, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return LaunchTask{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var t LaunchTask
	err = db.QueryRow(c.CTX,
		`UPDATE launch_task SET completed = $2 WHERE id = $1
		 RETURNING id, service_id, completed`,
		taskID, completed,
	).Scan(&t.ID, &t.ServiceID, &t.Completed)
	if err != nil {
		return LaunchTask{}, logs.Error(err)
	}
	return t, nil
}

func (c CMS) deleteLaunchTask(taskID int) error {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	_, err = db.Exec(c.CTX, "DELETE FROM launch_task WHERE id = $1", taskID)
	if err != nil {
		return logs.Error(err)
	}
	return nil
}

func (c CMS) AllowedOrigins() ([]string, error) {
	origins := make([]string, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return origins, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT non_url, alternatives FROM services")
	if err != nil {
		return nil, logs.Error(err)
	}
	type service struct {
		URL  *string
		Alts *string
	}

	defer rows.Close()
	for rows.Next() {
		var s service
		if err := rows.Scan(&s.URL, &s.Alts); err != nil {
			return nil, logs.Error(err)
		}
		if s.Alts != nil {
			alts := strings.Split(*s.Alts, ",")
			for _, alt := range alts {
				if alt != "" {
					origins = append(origins, fmt.Sprintf("https://%s", alt), fmt.Sprintf("https://www.%s", alt))
				}
			}
		}
		origins = append(origins, fmt.Sprintf("https://%s", *s.URL), fmt.Sprintf("https://www.%s", *s.URL))
	}

	return origins, nil
}
