package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"
)

func TestLoad(t *testing.T) {
	type args struct {
		serverConfigFileFromFlag string
	}
	tests := []struct {
		name    string
		args    args
		want    *Frontier
		wantErr bool
	}{
		{
			name: "load yaml file successfully while parsing time.Duration value",
			args: args{
				serverConfigFileFromFlag: "testdata/use_duration.yaml",
			},
			want: &Frontier{
				Version: 1,
				DB: db.Config{
					Driver:          "postgres",
					URL:             "postgres://frontier:@localhost:5432/frontier?sslmode=disable",
					MaxIdleConns:    10,
					MaxOpenConns:    10,
					ConnMaxLifeTime: time.Duration(10) * time.Millisecond,
					MaxQueryTimeout: time.Duration(500) * time.Millisecond,
				},
				SpiceDB: spicedb.Config{
					Host:         "spicedb.localhost",
					Port:         "50051",
					PreSharedKey: "randomkey",
					Consistency:  spicedb.ConsistencyLevelBestEffort.String(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.serverConfigFileFromFlag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Version, tt.want.Version) {
				t.Errorf("Load() got = %v\nwant %v", got.Version, tt.want.Version)
			}
			if !reflect.DeepEqual(got.DB, tt.want.DB) {
				t.Errorf("Load() got = %v\nwant %v", got.DB, tt.want.DB)
			}
			if !reflect.DeepEqual(got.SpiceDB, tt.want.SpiceDB) {
				t.Errorf("Load() got = %v\nwant %v", got.SpiceDB, tt.want.SpiceDB)
			}
		})
	}
}

func TestLoadDeletion(t *testing.T) {
	t.Run("reads namespace keys, rules and the user delete email", func(t *testing.T) {
		want := deleter.Config{
			Enabled: true,
			Resources: map[string]deleter.Rules{
				"compute/machine": {
					deleter.ParentProject:      deleter.RuleBlock,
					deleter.ParentOrganization: deleter.RuleAllow,
				},
				"compute/order": {
					deleter.ParentProject:      deleter.RuleBlockUnlessDeletable,
					deleter.ParentOrganization: deleter.RuleBlockUnlessDeletable,
				},
				"storage/file": {
					deleter.ParentOrganization: deleter.RuleAllow,
				},
			},
			UserDeleteNotice: deleter.UserDeleteNoticeConfig{
				Subject: "Your account was deleted",
				Body:    "<p>Hi {{.User.Title}}</p>",
			},
		}

		got, err := Load("testdata/deletion.yaml")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !reflect.DeepEqual(got.App.Deletion, want) {
			t.Errorf("Load() got = %v\nwant %v", got.App.Deletion, want)
		}
	})

	t.Run("an environment variable turns the check off", func(t *testing.T) {
		t.Setenv("FRONTIER_SERVICE_APP_DELETION_ENABLED", "false")

		got, err := Load("testdata/deletion.yaml")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got.App.Deletion.Enabled {
			t.Errorf("Load() got enabled = true, want false")
		}
	})

	t.Run("a misspelled key reaches the boot check", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		body := "app:\n  deletion:\n    enabled: true\n    resources:\n      compute/machine: { projct: block }\n"
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}

		got, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if err := got.App.Deletion.Validate(); err == nil || !strings.Contains(err.Error(), `unknown key "projct"`) {
			t.Errorf("Validate() error = %v, want an unknown key error", err)
		}
	})
}
