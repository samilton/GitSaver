package github

type WebhookPayload struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Sender     User       `json:"sender"`
	Created    bool       `json:"created"`
	Deleted    bool       `json:"deleted"`
	Forced     bool       `json:"forced"`
	BaseRef    *string    `json:"base_ref"`
	Compare    string     `json:"compare"`
	Commits    []Commit   `json:"commits"`
	HeadCommit Commit     `json:"head_commit"`
}

type Repository struct {
	ID              int64    `json:"id"`
	NodeID          string   `json:"node_id"`
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Private         bool     `json:"private"`
	Owner           User     `json:"owner"`
	HTMLURL         string   `json:"html_url"`
	Description     *string  `json:"description"`
	Fork            bool     `json:"fork"`
	URL             string   `json:"url"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	PushedAt        int64    `json:"pushed_at"`
	GitURL          string   `json:"git_url"`
	SSHURL          string   `json:"ssh_url"`
	CloneURL        string   `json:"clone_url"`
	Size            int      `json:"size"`
	StargazersCount int      `json:"stargazers_count"`
	WatchersCount   int      `json:"watchers_count"`
	Language        string   `json:"language"`
	HasIssues       bool     `json:"has_issues"`
	HasProjects     bool     `json:"has_projects"`
	HasDownloads    bool     `json:"has_downloads"`
	HasWiki         bool     `json:"has_wiki"`
	HasPages        bool     `json:"has_pages"`
	HasDiscussions  bool     `json:"has_discussions"`
	ForksCount      int      `json:"forks_count"`
	Archived        bool     `json:"archived"`
	Disabled        bool     `json:"disabled"`
	OpenIssuesCount int      `json:"open_issues_count"`
	License         *string  `json:"license"`
	AllowForking    bool     `json:"allow_forking"`
	IsTemplate      bool     `json:"is_template"`
	Topics          []string `json:"topics"`
	Visibility      string   `json:"visibility"`
	Forks           int      `json:"forks"`
	OpenIssues      int      `json:"open_issues"`
	Watchers        int      `json:"watchers"`
	DefaultBranch   string   `json:"default_branch"`
	MasterBranch    string   `json:"master_branch"`
}

type User struct {
	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commit struct {
	ID        string   `json:"id"`
	TreeID    string   `json:"tree_id"`
	Distinct  bool     `json:"distinct"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    Author   `json:"author"`
	Committer Author   `json:"committer"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
