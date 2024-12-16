package config

import (
	"reflect"
	"testing"
	"time"

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
