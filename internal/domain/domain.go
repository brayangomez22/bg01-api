// Package domain defines the JSON contract exposed by the API and emitted by
// /export. These types mirror the frontend's src/types/domain.ts (camelCase
// JSON tags) — when the frontend contract changes, these must follow.
package domain

// MediaRef is an image/media reference. Dimensions are required to prevent CLS.
type MediaRef struct {
	Src         string `json:"src"`
	Alt         string `json:"alt"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Placeholder string `json:"placeholder,omitempty"`
}

// LabelledValue backs stats and metrics.
type LabelledValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Period is a date range. End is optional for ongoing missions; for experience
// it is always present and may be the literal "present".
type Period struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

// Pilot is the profile singleton.
type Pilot struct {
	Name      string          `json:"name"`
	Callsign  string          `json:"callsign"`
	Role      string          `json:"role"`
	Available bool            `json:"available"`
	Location  string          `json:"location"`
	Bio       string          `json:"bio"`
	Manifesto string          `json:"manifesto"`
	Stats     []LabelledValue `json:"stats"`
	Avatar    MediaRef        `json:"avatar"`
	ResumeURL string          `json:"resumeUrl"`
}

// MissionLinks holds optional external references for a mission.
type MissionLinks struct {
	Live      string `json:"live,omitempty"`
	Repo      string `json:"repo,omitempty"`
	CaseStudy string `json:"caseStudy,omitempty"`
}

// Mission is a project.
type Mission struct {
	ID            string          `json:"id"`
	Code          string          `json:"code"`
	Title         string          `json:"title"`
	Summary       string          `json:"summary"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	Role          string          `json:"role"`
	DurationLabel string          `json:"durationLabel"`
	Period        Period          `json:"period"`
	Technologies  []string        `json:"technologies"`
	Highlights    []string        `json:"highlights"`
	Challenges    []string        `json:"challenges"`
	Metrics       []LabelledValue `json:"metrics,omitempty"`
	Links         MissionLinks    `json:"links"`
	Cover         MediaRef        `json:"cover"`
	Gallery       []MediaRef      `json:"gallery,omitempty"`
	Featured      bool            `json:"featured"`
	Order         int             `json:"order"`
}

// TechnologyPlanet feeds the orbital visualization.
type TechnologyPlanet struct {
	Color string `json:"color"`
	Size  string `json:"size"`
	Orbit int    `json:"orbit"`
}

// Technology is a stack "planet". UsedInMissions is derived in /export.
type Technology struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Category       string           `json:"category"`
	Proficiency    int              `json:"proficiency"`
	Since          string           `json:"since"`
	Description    string           `json:"description"`
	UsedInMissions []string         `json:"usedInMissions,omitempty"`
	Planet         TechnologyPlanet `json:"planet"`
	Featured       bool             `json:"featured"`
	Order          int              `json:"order"`
}

// Experience is a logbook entry.
type Experience struct {
	ID               string   `json:"id"`
	Period           Period   `json:"period"`
	Company          string   `json:"company"`
	Role             string   `json:"role"`
	Location         string   `json:"location"`
	Summary          string   `json:"summary"`
	Responsibilities []string `json:"responsibilities"`
	Achievements     []string `json:"achievements"`
	Technologies     []string `json:"technologies"`
	Type             string   `json:"type"`
	Order            int      `json:"order"`
}

// TrainingSim is a curated practice repo. Order is admin-only (not in the
// frontend contract) and drives list ordering.
type TrainingSim struct {
	Repo    string   `json:"repo"`
	Code    string   `json:"code"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Stack   []string `json:"stack"`
	Year    int      `json:"year"`
	RepoURL string   `json:"repoUrl"`
	Order   int      `json:"order"`
}

// ArchiveSection is a shelf in the knowledge archive.
type ArchiveSection struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Label string `json:"label"`
	Order int    `json:"order"`
}

// ArchiveSegment is one structured body block of an archive record.
type ArchiveSegment struct {
	Kind  string   `json:"kind"`            // h | p | list | code
	Text  string   `json:"text,omitempty"`  // h, p
	Items []string `json:"items,omitempty"` // list
	Lang  string   `json:"lang,omitempty"`  // code
	Code  string   `json:"code,omitempty"`  // code
}

// ArchiveRecord is a knowledge entry (meta + body).
type ArchiveRecord struct {
	ID             string           `json:"id"`
	Code           string           `json:"code"`
	Title          string           `json:"title"`
	Abstract       string           `json:"abstract"`
	Section        string           `json:"section"`
	Tags           []string         `json:"tags"`
	ArchivedAt     string           `json:"archivedAt"`
	ReadingMinutes int              `json:"readingMinutes"`
	Body           []ArchiveSegment `json:"body"`
	Refs           []string         `json:"refs"`
	Order          int              `json:"order"`
}

// Frequency is a contact channel. Order is admin-only and drives ordering.
type Frequency struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Handle  string `json:"handle"`
	URL     string `json:"url"`
	Icon    string `json:"icon"`
	Primary bool   `json:"primary"`
	Order   int    `json:"order"`
}

// SiteCopyEntry is one editable free-text string (admin CRUD form).
type SiteCopyEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Export is the full content bundle consumed by the build-time export.
type Export struct {
	Pilot        Pilot             `json:"pilot"`
	Missions     []Mission         `json:"missions"`
	Technologies []Technology      `json:"technologies"`
	Experience   []Experience      `json:"experience"`
	Training     []TrainingSim     `json:"training"`
	Archive      ArchiveBundle     `json:"archive"`
	Frequencies  []Frequency       `json:"frequencies"`
	SiteCopy     map[string]string `json:"siteCopy"`
}

// ArchiveBundle groups the archive shelves and their records.
type ArchiveBundle struct {
	Sections []ArchiveSection `json:"sections"`
	Records  []ArchiveRecord  `json:"records"`
}
