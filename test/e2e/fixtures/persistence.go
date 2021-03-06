package fixtures

import (
	"k8s.io/client-go/kubernetes"
	"upper.io/db.v3/lib/sqlbuilder"

	"github.com/argoproj/argo/config"
	"github.com/argoproj/argo/persist/sqldb"
)

type Persistence struct {
	session               sqlbuilder.Database
	offloadNodeStatusRepo sqldb.OffloadNodeStatusRepo
	workflowArchive       sqldb.WorkflowArchive
}

func newPersistence(kubeClient kubernetes.Interface) *Persistence {
	configController := config.NewController(Namespace, "workflow-controller-configmap", kubeClient)
	wcConfig, err := configController.Get()
	if err != nil {
		panic(err)
	}
	persistence := wcConfig.Persistence
	if persistence != nil {
		if persistence.PostgreSQL != nil {
			persistence.PostgreSQL.Host = "localhost"
		}
		if persistence.MySQL != nil {
			persistence.MySQL.Host = "localhost"
		}
		session, tableName, err := sqldb.CreateDBSession(kubeClient, Namespace, persistence)
		if err != nil {
			panic(err)
		}
		offloadNodeStatusRepo, err := sqldb.NewOffloadNodeStatusRepo(session, persistence.GetClusterName(), tableName)
		if err != nil {
			panic(err)
		}
		workflowArchive := sqldb.NewWorkflowArchive(session, persistence.GetClusterName(), Namespace, wcConfig.InstanceID)
		return &Persistence{session, offloadNodeStatusRepo, workflowArchive}
	} else {
		return &Persistence{offloadNodeStatusRepo: sqldb.ExplosiveOffloadNodeStatusRepo, workflowArchive: sqldb.NullWorkflowArchive}
	}
}

func (s *Persistence) IsEnabled() bool {
	return s.offloadNodeStatusRepo.IsEnabled()
}

func (s *Persistence) Close() {
	if s.IsEnabled() {
		err := s.session.Close()
		if err != nil {
			panic(err)
		}
	}
}
