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
	ID          int           `json:"-"`
	Name        string        `json:"name"`
	SearchName  string        `json:"searchName"`
	Description string        `json:"description"`
	Status      string        `json:"status"`
	URL         string        `json:"url"`
	LaunchDate  LaunchDate    `json:"launchDate"`
	Progress    int           `json:"progress"`
	Progress2   float32       `json:"progress2"`
	Icon        string        `json:"icon"`
	FullDesc    string        `json:"fullDesc"`
	Uptime      string        `json:"uptime"`
	Launched    bool          `json:"launched"`
	Links       []ProjectLink `json:"links,omitempty"`
	Milestones  []Milestone   `json:"milestones,omitempty"`
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

type Milestone struct {
	ID            int     `json:"id,omitempty"`
	ServiceID     int     `json:"serviceId,omitempty"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Category      string  `json:"category"`
	Status        string  `json:"status"`
	TargetDate    *string `json:"targetDate,omitempty"`
	CompletedDate *string `json:"completedDate,omitempty"`
	SortOrder     int     `json:"sortOrder"`
}

type CreateServiceRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	FullDesc    string     `json:"fullDesc"`
	Status      string     `json:"status"`
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

type CreateMilestoneRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	Status        string  `json:"status"`
	TargetDate    *string `json:"targetDate"`
	CompletedDate *string `json:"completedDate"`
	SortOrder     int     `json:"sortOrder"`
}

type UpdateMilestoneRequest struct {
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	Category      *string `json:"category,omitempty"`
	Status        *string `json:"status,omitempty"`
	TargetDate    *string `json:"targetDate,omitempty"`
	CompletedDate *string `json:"completedDate,omitempty"`
	SortOrder     *int    `json:"sortOrder,omitempty"`
}

func NewCMS(config *ConfigBuilder.Config, flags *flagsService.Client, errChan chan error) *CMS {
	return &CMS{
		Config:       config,
		Flags:        flags,
		CTX:          context.Background(),
		ErrorChannel: errChan,
	}
}

func calculateProgressFromStatuses(statuses []string) float32 {
	totalRows := 0
	completedRows := 0

	for _, status := range statuses {
		if status == "cancelled" {
			continue
		}
		totalRows++
		if status == "completed" {
			completedRows++
		}
	}

	if totalRows == 0 {
		return 0
	}

	return (float32(completedRows) / float32(totalRows)) * 100
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

	rows, err := db.Query(c.CTX, "SELECT id, name, search_name, description, status, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime, live FROM services WHERE started = true")
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
			&service.Status,
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

	if err := db.QueryRow(c.CTX, "SELECT id, name, search_name, description, status, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime, live FROM services WHERE search_name = $1", name).Scan(
		&service.ID,
		&service.Name,
		&service.SearchName,
		&service.Description,
		&service.Status,
		&service.LaunchDate.Year,
		&service.LaunchDate.Month,
		&service.LaunchDate.Day,
		&service.URL,
		&service.Progress,
		&service.Icon,
		&service.FullDesc,
		&service.Uptime,
		&service.Launched); err != nil {
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

	milestones, err := c.getServiceMilestones(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Milestones = milestones

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
	rows, err := db.Query(c.CTX, "SELECT status FROM project_milestones WHERE service_id = $1", id)
	if err != nil {
		return 0, logs.Error(err)
	}
	statuses := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return 0, logs.Error(err)
		}
		statuses = append(statuses, status)
	}

	return calculateProgressFromStatuses(statuses), nil
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

func (c CMS) getServiceMilestones(serviceID int) ([]Milestone, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, `SELECT id, service_id, title, COALESCE(description, ''), category, status, target_date::text, completed_date::text, sort_order
		FROM project_milestones
		WHERE service_id = $1
		ORDER BY sort_order, id`, serviceID)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var milestones []Milestone
	for rows.Next() {
		var item Milestone
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.Title, &item.Description, &item.Category, &item.Status, &item.TargetDate, &item.CompletedDate, &item.SortOrder); err != nil {
			return nil, logs.Error(err)
		}
		milestones = append(milestones, item)
	}

	return milestones, nil
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
	status := req.Status
	if status == "" {
		status = "planned"
	}
	var svc Service
	err = db.QueryRow(c.CTX,
		`INSERT INTO services (name, search_name, description, full_description, status, url, icon, uptime, launch_year, launch_month, launch_day, started)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true)
		 RETURNING id, name, search_name, description, status, url, launch_year, launch_month, launch_day, progress, icon, full_description, uptime, live`,
		req.Name, searchName, req.Description, req.FullDesc, status, req.URL, req.Icon, req.Uptime,
		req.LaunchDate.Year, req.LaunchDate.Month, req.LaunchDate.Day,
	).Scan(&svc.ID, &svc.Name, &svc.SearchName, &svc.Description, &svc.Status, &svc.URL,
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
		`UPDATE services SET name=$1, description=$2, full_description=$3, status=$4, url=$5, icon=$6, uptime=$7,
		 launch_year=$8, launch_month=$9, launch_day=$10
		 WHERE search_name=$11
		 RETURNING id, name, search_name, description, status, url, launch_year, launch_month, launch_day, progress, icon, full_description, uptime, live`,
		req.Name, req.Description, req.FullDesc, req.Status, req.URL, req.Icon, req.Uptime,
		req.LaunchDate.Year, req.LaunchDate.Month, req.LaunchDate.Day, name,
	).Scan(&svc.ID, &svc.Name, &svc.SearchName, &svc.Description, &svc.Status, &svc.URL,
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

func (c CMS) getMilestones(serviceName string) ([]Milestone, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, `SELECT pm.id, pm.service_id, pm.title, COALESCE(pm.description, ''), pm.category, pm.status, pm.target_date::text, pm.completed_date::text, pm.sort_order
		FROM project_milestones pm
		JOIN services s ON s.id = pm.service_id
		WHERE s.search_name = $1
		ORDER BY pm.sort_order, pm.id`, serviceName)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var milestones []Milestone
	for rows.Next() {
		var item Milestone
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.Title, &item.Description, &item.Category, &item.Status, &item.TargetDate, &item.CompletedDate, &item.SortOrder); err != nil {
			return nil, logs.Error(err)
		}
		milestones = append(milestones, item)
	}
	return milestones, nil
}

func (c CMS) createMilestone(serviceName string, req CreateMilestoneRequest) (Milestone, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return Milestone{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var item Milestone
	err = db.QueryRow(c.CTX,
		`INSERT INTO project_milestones (service_id, title, description, category, status, target_date, completed_date, sort_order)
		 SELECT id, $2, NULLIF($3, ''), $4, $5, $6::date, $7::date, $8 FROM services WHERE search_name = $1
		 RETURNING id, service_id, title, COALESCE(description, ''), category, status, target_date::text, completed_date::text, sort_order`,
		serviceName, req.Title, req.Description, req.Category, req.Status, req.TargetDate, req.CompletedDate, req.SortOrder,
	).Scan(&item.ID, &item.ServiceID, &item.Title, &item.Description, &item.Category, &item.Status, &item.TargetDate, &item.CompletedDate, &item.SortOrder)
	if err != nil {
		return Milestone{}, logs.Error(err)
	}
	return item, nil
}

func (c CMS) updateMilestone(itemID int, req UpdateMilestoneRequest) (Milestone, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return Milestone{}, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	var item Milestone
	err = db.QueryRow(c.CTX,
		`UPDATE project_milestones SET
		 title = COALESCE($2, title),
		 description = COALESCE(NULLIF($3, ''), description),
		 category = COALESCE($4, category),
		 status = COALESCE($5, status),
		 target_date = COALESCE($6::date, target_date),
		 completed_date = COALESCE($7::date, completed_date),
		 sort_order = COALESCE($8, sort_order),
		 updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, service_id, title, COALESCE(description, ''), category, status, target_date::text, completed_date::text, sort_order`,
		itemID, req.Title, req.Description, req.Category, req.Status, req.TargetDate, req.CompletedDate, req.SortOrder,
	).Scan(&item.ID, &item.ServiceID, &item.Title, &item.Description, &item.Category, &item.Status, &item.TargetDate, &item.CompletedDate, &item.SortOrder)
	if err != nil {
		return Milestone{}, logs.Error(err)
	}
	return item, nil
}

func (c CMS) deleteMilestone(itemID int) error {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	_, err = db.Exec(c.CTX, "DELETE FROM project_milestones WHERE id = $1", itemID)
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
