/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"database/sql"
	// "fmt"
	// "io"
	// "strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	dbv1 "github.com/lovelinuxalot/postgresql-operator/api/v1"
	"github.com/lovelinuxalot/postgresql-operator/internal/postgres"
)

var postgresDatabaseFinalizer = "postgresdatabase.finalizers.my.domain"

// PostgresDatabaseReconciler reconciles a PostgresDatabase object
type PostgresDatabaseReconciler struct {
	Client client.Client
	Scheme  *runtime.Scheme
	Log     logr.Logger
}

// +kubebuilder:rbac:groups=db.pandarocks.com,resources=postgresdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=db.pandarocks.com,resources=postgresdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=db.pandarocks.com,resources=postgresdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgresDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *PostgresDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	pgdb := &dbv1.PostgresDatabase{}
	err := r.Client.Get(ctx, req.NamespacedName, pgdb)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Connect to PostgreSQL
	db, err := postgres.Connect()
	if err != nil {
		return ctrl.Result{}, err
	}
	defer db.Close()


	// Create the database if it doesn't exist
	err = postgres.CreateDatabase(ctx, db, pgdb.ObjectMeta.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pgdb.Spec.DropOnDelete {
		// Handle deletion logic
		if err := r.handleDeletion(ctx, db, pgdb); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// Ensure the finalizer is set so we can handle cleanup
		if !containsString(pgdb.GetFinalizers(), postgresDatabaseFinalizer) {
			pgdb.SetFinalizers(append(pgdb.GetFinalizers(), postgresDatabaseFinalizer))
			if err := r.Client.Update(ctx, pgdb); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1.PostgresDatabase{}).
		Complete(r)
}

func (r *PostgresDatabaseReconciler) handleDeletion(ctx context.Context, db *sql.DB, pgdb *dbv1.PostgresDatabase) error {
	if containsString(pgdb.GetFinalizers(), postgresDatabaseFinalizer) {
		// Drop the database
		err := postgres.DropDatabase(ctx, db, pgdb.ObjectMeta.Name)
		if err != nil {
			return err
		}

		// Remove our finalizer from the list and update it
		pgdb.SetFinalizers(removeString(pgdb.GetFinalizers(), postgresDatabaseFinalizer))
		if err := r.Client.Update(ctx, pgdb); err != nil {
			return err
		}
	}
	return nil
}

// Helper functions to manage finalizers

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
