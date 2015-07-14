package marathon

import (
	"encoding/json"
	"github.com/QubitProducts/bamboo/configuration"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"log"
	"time"
)

// Describes an app process running
type Task struct {
	Host       string
	Port       int
	SecondPort int
}

// An app may have multiple processes
type App struct {
	Id              string
	EscapedId       string
	HealthCheckPath string
	Tasks           []Task
	ServicePort     int
	Env             map[string]string
	ResourcePath	string
}

type AppList []App

func (slice AppList) Len() int {
	return len(slice)
}

func (slice AppList) Less(i, j int) bool {
	return slice[i].Id < slice[j].Id
}

func (slice AppList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type MarathonTaskList []MarathonTask

type MarathonTasks struct {
	Tasks MarathonTaskList `json:tasks`
}

type HealthCheckResult struct {
	TaskId              string
	FirstSuccess        string
	LastSuccess         string
	LastFailure         string
	ConsecutiveFailures int
	Alive               bool

}

type MarathonTask struct {
	AppId              string
	Id                 string
	Host               string
	Ports              []int
	ServicePorts       []int
	StartedAt          string
	StagedAt           string
	Version            string
	HealthCheckResults []HealthCheckResult
}

func (slice MarathonTaskList) Len() int {
	return len(slice)
}

func (slice MarathonTaskList) Less(i, j int) bool {
	return slice[i].StagedAt < slice[j].StagedAt
}

func (slice MarathonTaskList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type MarathonApps struct {
	Apps []MarathonApp `json:apps`
}

type MarathonApp struct {
	Id           string            `json:id`
	HealthChecks []HealthChecks    `json:healthChecks`
	Ports        []int             `json:ports`
	Env          map[string]string `json:env`
	Labels       map[string]string `json:labels`
}

type HealthChecks struct {
	Path string `json:path`
}

func fetchMarathonApps(endpoint string) (map[string]MarathonApp, error) {
	response, err := http.Get(endpoint + "/v2/apps")

	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		var appResponse MarathonApps

		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(contents, &appResponse)
		if err != nil {
			return nil, err
		}

		dataById := map[string]MarathonApp{}

		for _, appConfig := range appResponse.Apps {
			dataById[appConfig.Id] = appConfig

		}

		return dataById, nil
	}
}

func fetchTasks(endpoint string) (map[string][]MarathonTask, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/v2/tasks", nil)
	req.Header.Add("Accept", "application/json")
	response, err := client.Do(req)

	var tasks MarathonTasks
	if err != nil {
		return nil, err
	} else {
		contents, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(contents, &tasks)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		taskList := tasks.Tasks
		sort.Sort(taskList)
		tasksById := map[string][]MarathonTask{}
		for _, task := range taskList {
			//TODO; check the alive

			if (task.HealthCheckResults !=  nil) {

				lastSuccessTime, errSuccess := time.Parse("2006-01-02T15:04:05Z07:00", task.HealthCheckResults[0].LastSuccess)
				lastFailureTime, errFailure := time.Parse("2006-01-02T15:04:05Z07:00", task.HealthCheckResults[0].LastFailure)

				if (errSuccess == nil) {
					log.Println("There was a successful healthcheck")

					if (errFailure == nil) {
						if (lastSuccessTime.After(lastFailureTime)) {
							if tasksById[task.AppId] == nil {
								tasksById[task.AppId] = []MarathonTask{}
							}
							tasksById[task.AppId] = append(tasksById[task.AppId], task)
						}
					} else {
						if tasksById[task.AppId] == nil {
							tasksById[task.AppId] = []MarathonTask{}
						}
						tasksById[task.AppId] = append(tasksById[task.AppId], task)
					}
				} else {
					log.Println("Nothing to update here-----------")
				}
			} else {
				log.Println("There is no healthcheck on this app")
				if tasksById[task.AppId] == nil {
					tasksById[task.AppId] = []MarathonTask{}
				}
				tasksById[task.AppId] = append(tasksById[task.AppId], task)
			}

		}

		log.Printf("I have tasks of length: %v", len(tasksById))

		return tasksById, nil
	}
}

func createApps(tasksById map[string][]MarathonTask, marathonApps map[string]MarathonApp) AppList {
	log.Println("ENTER >> createApps")
	apps := AppList{}

	for appId, tasks := range tasksById {
		simpleTasks := []Task{}

		for _, task := range tasks {
			if (len(task.Ports) > 0) {
				//check if the app has health checks
				if (len(marathonApps[appId].HealthChecks) > 0) {
					log.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ The app has healthckechs: %v", len(marathonApps[appId].HealthChecks))
					//check if the app has healthckeck results
					if (len(task.HealthCheckResults) > 0) {
						log.Println("^^^^^^^^^^^^^^^^^&^^^^^^^^^^^^^task has healthckeck results %v", len(task.HealthCheckResults))

						//Check if the healthcheck results have passed or not
						for _, healthCheckResult := range task.HealthCheckResults {
							lastSuccessTime, errSuccess := time.Parse("2006-01-02T15:04:05Z07:00", healthCheckResult.LastSuccess)
							lastFailureTime, errFailure := time.Parse("2006-01-02T15:04:05Z07:00", healthCheckResult.LastFailure)
							//check last success
							if (errSuccess == nil) {
								log.Println("There was a successful healthcheck")
								//check if there were failures
								if (errFailure == nil) {
									//check if the last success is newer than the last failure
									if (lastSuccessTime.After(lastFailureTime)) {
										//all ok here
										simpleTasks = append(simpleTasks, Task{Host: task.Host, Port: task.Ports[0], SecondPort: task.Ports[1]})
									}
									//there were no failures
								} else {
									simpleTasks = append(simpleTasks, Task{Host: task.Host, Port: task.Ports[0], SecondPort: task.Ports[1]})
								}
							}
						}
						//there are no healtheck results
					}
					//the app has no healthcecks
				} else {
					simpleTasks = append(simpleTasks, Task{Host: task.Host, Port: task.Ports[0], SecondPort: task.Ports[1]})
				}

			}
		}

		// Try to handle old app id format without slashes
		appPath := appId
		if !strings.HasPrefix(appId, "/") {
			appPath = "/" + appId
		}


		//check if there were any ttasks created for the app
		if (len(simpleTasks) > 0) {
			app := App{
				// Since Marathon 0.7, apps are namespaced with path
				Id: appPath,
				// Used for template
				EscapedId:       strings.Replace(appId, "/", "::", -1),
				Tasks:           simpleTasks,
				HealthCheckPath: parseHealthCheckPath(marathonApps[appId].HealthChecks),
				Env:             marathonApps[appId].Env,
				ResourcePath:	 parseResourcePath(marathonApps[appId], appId),
			}

			if len(marathonApps[appId].Ports) > 0 {
				app.ServicePort = marathonApps[appId].Ports[0]
			}

			if (len(marathonApps[appId].HealthChecks) > 0) {
				log.Println("The app has healthchecks")
			}

			apps = append(apps, app)
		}
	}

	log.Println("EXIT << createApps")
	return apps
}

func parseHealthCheckPath(checks []HealthChecks) string {
	if len(checks) > 0 {
		return checks[0].Path
	}
	return ""
}

func parseResourcePath(marathonApp MarathonApp, appId string) string {
	if (marathonApp.Labels["resourcePath"] != ""){
		return marathonApp.Labels["resourcePath"]
	}

	return appId
}



/*
	Apps returns a struct that describes Marathon current app and their
	sub tasks information.

	Parameters:
		endpoint: Marathon HTTP endpoint, e.g. http://localhost:8080
*/
func FetchApps(maraconf configuration.Marathon) (AppList, error) {
	var applist AppList
	var err error

	// try all configured endpoints until one succeeds
	for _, url := range maraconf.Endpoints() {
		applist, err = _fetchApps(url)
		if err == nil {
			return applist, err
		}
	}
	// return last error
	return nil, err
}

func _fetchApps(url string) (AppList, error) {
	log.Println("ENTER _fetchApps")
	tasks, err := fetchTasks(url)

	if err != nil {
		return nil, err
	}

	marathonApps, err := fetchMarathonApps(url)
	if err != nil {
		return nil, err
	}

	apps := createApps(tasks, marathonApps)
	sort.Sort(apps)

	log.Println("EXIT _fetchApps")

	return apps, nil
}
