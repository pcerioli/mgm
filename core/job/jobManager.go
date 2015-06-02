package job

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager manages jobs, updating database, and notifying subscribed parties
type Manager interface {
	Subscribe() core.Subscription
	FileUploaded(int, uuid.UUID, []byte)
	GetJobByID(id int) (mgm.Job, error)
	DeleteJob(job mgm.Job) error
	CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (mgm.Job, error)
	GetJobsForUser(userID uuid.UUID) ([]mgm.Job, error)
}

type fileUpload struct {
	JobID int
	User  uuid.UUID
	File  []byte
}

// NewManager constructs a jobManager for use
func NewManager(filePath string, db database.Database, log logger.Log) Manager {

	subscribeChan := make(chan chan<- mgm.Job, 32)
	unsubscribeChan := make(chan chan<- mgm.Job, 32)
	notifyChan := make(chan mgm.Job, 32)

	j := jobMgr{}
	j.fileUp = make(chan fileUpload, 32)
	j.localPath = filePath
	j.log = logger.Wrap("JOB", log)
	j.subscribe = subscribeChan
	j.unsubscribe = unsubscribeChan
	j.broadcast = notifyChan
	j.db = jobDatabase{db}

	j.subs = core.NewSubscriptionManager()

	go j.process()

	return j
}

type jobMgr struct {
	fileUp      chan fileUpload
	subscribe   chan chan<- mgm.Job
	unsubscribe chan chan<- mgm.Job
	broadcast   chan mgm.Job
	db          jobDatabase

	subs core.SubscriptionManager

	log logger.Log

	localPath string
}

func (jm jobMgr) FileUploaded(id int, user uuid.UUID, data []byte) {
	jm.fileUp <- fileUpload{id, user, data}
}

func (jm jobMgr) GetJobByID(id int) (mgm.Job, error) {
	return jm.db.GetJobByID(id)
}

func (jm jobMgr) DeleteJob(j mgm.Job) error {
	return jm.db.DeleteJob(j)
}

func (jm jobMgr) Subscribe() core.Subscription {
	return jm.subs.Subscribe()
}

func (jm jobMgr) GetJobsForUser(userID uuid.UUID) ([]mgm.Job, error) {
	return jm.db.GetJobsForUser(userID)
}

func (jm jobMgr) CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (mgm.Job, error) {
	return jm.db.CreateLoadIarJob(owner, inventoryPath)
}

func (jm jobMgr) process() {

	go func() {
		for {
			select {

			case s := <-jm.fileUp:
				jm.log.Info("File Upload Received for task %v", s.JobID)
				// look up job
				j, err := jm.db.GetJobByID(s.JobID)
				if err != nil {
					//anything could have happened, but the job doesn't seem to exist, drop file
					continue
				}

				//make sure uploader owns the job in question
				if s.User != j.User {
					jm.log.Info("Attempted upload to job %v by %v, owned by %v", j.ID, s.User, j.User)
					continue
				}

				jm.log.Info("Retrieved job %v for %v", j.Type, j.User)

				switch j.Type {
				case "load_iar":
					iarJob := LoadIarJob{}
					err := json.Unmarshal([]byte(j.Data), &iarJob)
					if err != nil {
						jm.log.Info("Error parsing Load Iar job: %v", err.Error())
					}

					iarJob.Filename = uuid.NewV4()

					err = ioutil.WriteFile(path.Join(jm.localPath, iarJob.Filename.String()), s.File, 0644)
					if err != nil {
						jm.log.Error("Error writing file: ", err)
					}

					iarJob.Status = "Iar upload to MGM"

					data, _ := json.Marshal(iarJob)
					j.Data = string(data)

					jm.db.UpdateJob(j)

					jm.broadcast <- j
				}
			}
		}
	}()

}
