package main

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nikolai5slo/ttlock2mqtt/controller"
	"github.com/nikolai5slo/ttlock2mqtt/credentials"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
	"github.com/nikolai5slo/ttlock2mqtt/mqtt"
	"github.com/nikolai5slo/ttlock2mqtt/server"
	"github.com/nikolai5slo/ttlock2mqtt/server/handlers"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
	ttlockapi "github.com/nikolai5slo/ttlock2mqtt/ttlock-api"
	"github.com/schollz/jsonstore"
)

type deps struct {
	cfg                Config
	ttlockService      ttlock.Service
	server             *server.Server
	handlers           *handlers.Handlers
	lockStorage        locks.Storage
	credentialsStorage credentials.Storage
	mqtt               *mqtt.HAMqtt
	controller         *controller.Controller
}

func (d *deps) buildConfig() error {
	if _, err := os.Stat(".env"); err == nil {
		err = cleanenv.ReadConfig(".env", &d.cfg)

		if err != nil {
			log.Printf("cant load .env file: %s", err)
		}
	}

	return cleanenv.ReadEnv(&d.cfg)
}

func (d *deps) buildTTLockService() error {
	ttlockClient, err := ttlockapi.NewClientWithResponses(d.cfg.TTLock.Server)

	if err != nil {
		return err
	}

	d.ttlockService, err = ttlock.New(
		ttlock.WithTTLockClient(ttlockClient),
		ttlock.WithClientSecret(d.cfg.TTLock.ClientID, d.cfg.TTLock.ClientSecret),
	)

	return err
}

func (d *deps) buildStorages() (err error) {
	// Credentials storage
	d.credentialsStorage, err = credentials.NewJsonStore(new(jsonstore.JSONStore), d.cfg.Storage.FilePath)

	if err != nil {
		return
	}

	// Lock store
	d.lockStorage, err = locks.NewJsonStore(new(jsonstore.JSONStore), d.cfg.Storage.FilePath)

	return
}

func (d *deps) buildHandlers() (err error) {
	d.handlers, err = handlers.New(
		handlers.WithLockStorage(d.lockStorage),
		handlers.WithCredentialsStorage(d.credentialsStorage),
		handlers.WithTTlockService(d.ttlockService),
	)
	return
}

func (d *deps) buildServer() (err error) {
	d.server, err = server.New(
		server.WithHandlers(d.handlers),
		server.WithAddress(d.cfg.Server.Address),
	)
	return
}

func (d *deps) buildMqtt() (err error) {
	d.mqtt, err = mqtt.New(
		mqtt.WithBroker(d.cfg.Mqtt.Broker),
		mqtt.WithClientID(d.cfg.Mqtt.ClientID),
		mqtt.WithCredentials(d.cfg.Mqtt.Username, d.cfg.Mqtt.Password),
	)
	return
}

func (d *deps) buildController() (err error) {
	d.controller, err = controller.New(
		controller.WithLockStorage(d.lockStorage),
		controller.WithCredentialsStorage(d.credentialsStorage),
		controller.WithMqtt(d.mqtt),
		controller.WithTTlockService(d.ttlockService),
		controller.WithRefreshRate(d.cfg.TTLock.RefreshInterval),
	)
	return
}

func buildDeps() (*deps, error) {
	d := &deps{}

	fList := []func() error{
		d.buildConfig,
		d.buildTTLockService,
		d.buildStorages,
		d.buildHandlers,
		d.buildServer,
		d.buildMqtt,
		d.buildController,
	}

	for _, f := range fList {
		if err := f(); err != nil {
			return d, err
		}
	}

	return d, nil
}

func main() {

	d, err := buildDeps()

	if err != nil {
		log.Panicf("unable to build dependencies: %s", err)
	}

	defer d.controller.Close()

	d.controller.StartAutoRefresh()

	err = d.server.Run()

	if err != nil {
		panic(err)
	}

	log.Printf("Server shutdown")
}
