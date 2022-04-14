package traefik_plugin_metrics

import (
	"context"
	"time"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/influxdata/influxdb-client-go/v2"
)

// Config holds the plugin configuration.
type Config struct {
	ClientIP     string 		`json:"clientIP,omitempty"`
	ClientBucket string 		`json:"clientbucket,omitempty"`
	ClientMeasurement string	`json:"clientmeasurement,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{
		ClientIP: "",
		ClientBucket: "",
		ClientMeasurement: "",
	}
}

// metrics is a metrics plugin.
type metrics struct {
	Name       string
	Next       http.Handler
	Config     *Config
}

// New creates and returns a plugin instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &metrics{
		Name:   name,
		Next:   next,
		Config: config,
	}, nil
}

func (h *metrics) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.logger().ServeHTTP(rw, req)
}

func (h *metrics) logger() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//创建数据库接口
		client := influxdb2.NewClient(h.Config.ClientIP, fmt.Sprintf("%s:%s","",""))
		//确定写入点
		writeAPI := client.WriteAPIBlocking("", h.Config.ClientBucket)
		
		//获取要写入的内容
		rec := httptest.NewRecorder()
		h.Next.ServeHTTP(rec, r)

		//确定写入内容
		p := influxdb2.NewPoint(h.Config.ClientMeasurement,
			map[string]string{"class": "response"},
        	map[string]interface{}{"STATUS": rec.Code, "HOST": r.Host},
        	time.Now())
		//写入
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			fmt.Printf("Write error: %s\n", err.Error())
		}
		//关闭接口
		client.Close()

		for k, vv := range rec.Header() {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		data := rec.Body.Bytes()

		w.WriteHeader(rec.Code)
		w.Write(data)
	})
}
