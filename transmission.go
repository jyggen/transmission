package transmission

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sort"
)

const (
	StatusPaused       = 0
	StatusWait         = 1
	StatusCheck        = 2
	StatusDownloadWait = 3
	StatusDownload     = 4
	StatisSeedWait     = 5
	StatusSeed         = 6
)

//TransmissionClient to talk to transmission
type TransmissionClient struct {
	apiclient ApiClient
}

type Command struct {
	Method    string    `json:"method,omitempty"`
	Arguments arguments `json:"arguments,omitempty"`
	Result    string    `json:"result,omitempty"`
}

type arguments struct {
	Fields       []string     `json:"fields,omitempty"`
	Torrents     Torrents     `json:"torrents,omitempty"`
	Ids          []int        `json:"ids,omitempty"`
	DeleteData   bool         `json:"delete-local-data,omitempty"`
	DownloadDir  string       `json:"download-dir,omitempty"`
	MetaInfo     string       `json:"metainfo,omitempty"`
	Filename     string       `json:"filename,omitempty"`
	TorrentAdded TorrentAdded `json:"torrent-added"`
	Paused       bool         `json:"paused,omitempty"`
	Location     string       `json:"location,omitempty"`
}

//TrackerStat struct for tracker stats.
type TrackerStat struct {
	Announce              string `json:"announce"`
	AnnounceState         int    `json:"announceState"`
	DownloadCount         int    `json:"downloadCount"`
	HasAnnounced          bool   `json:"hasAnnounced"`
	HasScraped            bool   `json:"hasScraped"`
	Host                  string `json:"host"`
	ID                    uint64 `json:"id"`
	IsBackup              bool   `json:"isBackup"`
	LastAnnouncePeerCount int    `json:"lastAnnouncePeerCount"`
	LastAnnounceResult    string `json:"lastAnnounceResult"`
	LastAnnounceStartTime int64  `json:"lastAnnounceStartTime"`
	LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded"`
	LastAnnounceTime      int64  `json:"lastAnnounceTime"`
	LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut"`
	LastScrapeResult      string `json:"lastScrapeResult"`
	LastScrapeStartTime   int64  `json:"lastScrapeStartTime"`
	LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded"`
	LastScrapeTime        int64  `json:"lastScrapeTime"`
	LastScrapeTimedOut    int64  `json:"lastScrapeTimedOut"`
	LeecherCount          int    `json:"leecherCount"`
	NextAnnounceTime      int64  `json:"nextAnnounceTime"`
	NextScrapeTime        int64  `json:"nextScrapeTime"`
	Scrape                string `json:"scrape"`
	ScrapeState           int    `json:"scrapeState"`
	SeederCount           int    `json:"seederCount"`
	Tier                  int    `json:"tier"`
}

//File struct for tracker stats.
type File struct {
	BytesCompleted int64
	Length         int64
	Name           string
}

//Torrent struct for torrents
type Torrent struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Status        int           `json:"status"`
	AddedDate     int           `json:"addedDate"`
	LeftUntilDone int64         `json:"leftUntilDone"`
	Eta           int           `json:"eta"`
	UploadRatio   float64       `json:"uploadRatio"`
	RateDownload  int           `json:"rateDownload"`
	RateUpload    int           `json:"rateUpload"`
	DownloadDir   string        `json:"downloadDir"`
	IsFinished    bool          `json:"isFinished"`
	PercentDone   float64       `json:"percentDone"`
	SeedRatioMode int           `json:"seedRatioMode"`
	HashString    string        `json:"hashString"`
	Error         int           `json:"error"`
	ErrorString   string        `json:"errorString"`
	TrackerStats  []TrackerStat `json:"trackerStats"`
	Files         []File        `json:"files"`
}

// Torrents represent []Torrent
type Torrents []Torrent

// sorting types
type (
	byID        Torrents
	byName      Torrents
	byAddedDate Torrents
)

func (t byID) Len() int           { return len(t) }
func (t byID) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byID) Less(i, j int) bool { return t[i].ID < t[j].ID }

func (t byName) Len() int           { return len(t) }
func (t byName) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byName) Less(i, j int) bool { return t[i].Name < t[j].Name }

func (t byAddedDate) Len() int           { return len(t) }
func (t byAddedDate) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byAddedDate) Less(i, j int) bool { return t[i].AddedDate < t[j].AddedDate }

// methods of 'Torrents' to sort by ID, Name or AddedDate
func (t Torrents) SortByID(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byID(t)))
		return
	}
	sort.Sort(byID(t))
}

func (t Torrents) SortByName(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byName(t)))
		return
	}
	sort.Sort(byName(t))
}

func (t Torrents) SortByAddedDate(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byAddedDate(t)))
		return
	}
	sort.Sort(byAddedDate(t))
}

//TorrentAdded data returning
type TorrentAdded struct {
	HashString string `json:"hashString"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
}

//New create new transmission torrent
func New(url string, username string, password string) TransmissionClient {
	apiclient := NewClient(url, username, password)
	tc := TransmissionClient{apiclient: apiclient}
	return tc
}

//GetTorrents get a list of torrents
func (ac *TransmissionClient) GetTorrents() (Torrents, error) {
	cmd, err := NewGetTorrentsCmd()

	out, err := ac.ExecuteCommand(cmd)
	if err != nil {
		return nil, err
	}

	return out.Arguments.Torrents, nil
}

//GetTorrent get a torrent by its ID
func (ac *TransmissionClient) GetTorrent(id int) (Torrent, error) {
	cmd, err := NewGetTorrentsCmd()

	cmd.Arguments.Ids = []int{id}

	out, err := ac.ExecuteCommand(cmd)
	if err != nil {
		return Torrent{}, err
	}

	if len(out.Arguments.Torrents) != 1 {
		return Torrent{}, errors.New("no results found")
	}

	return out.Arguments.Torrents[0], nil
}

//StartTorrent start the torrent
func (ac *TransmissionClient) StartTorrent(id int) (string, error) {
	return ac.sendSimpleCommand("torrent-start", id)
}

//StopTorrent start the torrent
func (ac *TransmissionClient) StopTorrent(id int) (string, error) {
	return ac.sendSimpleCommand("torrent-stop", id)
}

//VerifyTorrent verify the torrent
func (ac *TransmissionClient) VerifyTorrent(id int) (string, error) {
	return ac.sendSimpleCommand("torrent-verify", id)
}

func NewGetTorrentsCmd() (*Command, error) {
	cmd := &Command{}

	cmd.Method = "torrent-get"
	cmd.Arguments.Fields = []string{"id", "name", "hashString",
		"status", "addedDate", "leftUntilDone", "eta", "uploadRatio",
		"rateDownload", "rateUpload", "downloadDir", "isFinished",
		"percentDone", "seedRatioMode", "error", "errorString",
		"trackerStats", "files"}

	return cmd, nil
}

func NewAddCmd() (*Command, error) {
	cmd := &Command{}
	cmd.Method = "torrent-add"
	return cmd, nil
}

func NewAddCmdByMagnet(magnetLink string) (*Command, error) {
	cmd, _ := NewAddCmd()
	cmd.Arguments.Filename = magnetLink
	return cmd, nil
}

func NewAddCmdByURL(url string) (*Command, error) {
	cmd, _ := NewAddCmd()
	cmd.Arguments.Filename = url
	return cmd, nil
}

func NewAddCmdByFilename(filename string) (*Command, error) {
	cmd, _ := NewAddCmd()
	cmd.Arguments.Filename = filename
	return cmd, nil
}

func NewAddCmdByFile(file string) (*Command, error) {
	cmd, _ := NewAddCmd()

	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cmd.Arguments.MetaInfo = base64.StdEncoding.EncodeToString(fileData)

	return cmd, nil
}

func (cmd *Command) SetDownloadDir(dir string) {
	cmd.Arguments.DownloadDir = dir
}

func NewSetCmd(id int) (*Command, error) {
	cmd := &Command{}
	cmd.Method = "torrent-set"
	cmd.Arguments.Ids = []int{id}
	return cmd, nil
}

func NewDelCmd(id int, removeFile bool) (*Command, error) {
	cmd := &Command{}
	cmd.Method = "torrent-remove"
	cmd.Arguments.Ids = []int{id}
	cmd.Arguments.DeleteData = removeFile
	return cmd, nil
}

func (ac *TransmissionClient) ExecuteCommand(cmd *Command) (*Command, error) {
	out := &Command{}

	body, err := json.Marshal(cmd)
	if err != nil {
		return out, err
	}
	output, err := ac.apiclient.Post(string(body))
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(output, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (ac *TransmissionClient) ExecuteAddCommand(addCmd *Command) (TorrentAdded, error) {
	outCmd, err := ac.ExecuteCommand(addCmd)
	if err != nil {
		return TorrentAdded{}, err
	}
	return outCmd.Arguments.TorrentAdded, nil
}

func encodeFile(file string) (string, error) {
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(fileData), nil
}

func (ac *TransmissionClient) sendSimpleCommand(method string, id int) (result string, err error) {
	cmd := Command{Method: method}
	cmd.Arguments.Ids = []int{id}
	resp, err := ac.sendCommand(cmd)
	return resp.Result, err
}

func (ac *TransmissionClient) sendCommand(cmd Command) (response Command, err error) {
	body, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	output, err := ac.apiclient.Post(string(body))
	if err != nil {
		return
	}
	err = json.Unmarshal(output, &response)
	if err != nil {
		return
	}
	return response, nil
}
