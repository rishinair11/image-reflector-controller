/*
Copyright 2020 The Flux authors

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

package controllers

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluxcd/pkg/runtime/controller"
	"github.com/fluxcd/pkg/runtime/testenv"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	ircbadger "github.com/fluxcd/image-reflector-controller/internal/database/badger"
	// +kubebuilder:scaffold:imports
)

// These tests make use of plain Go using Gomega for assertions.
// At the beginning of every (sub)test Gomega can be initialized
// using gomega.NewWithT.
// Refer to http://onsi.github.io/gomega/ to learn more about
// Gomega.

// for Eventually
const (
	timeout                = time.Second * 30
	contextTimeout         = time.Second * 20
	interval               = time.Second * 1
	reconciliationInterval = time.Second * 2
)

var (
	testEnv      *testenv.Environment
	testBadgerDB *badger.DB
	ctx          = ctrl.SetupSignalHandler()
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestMain(m *testing.M) {

	utilruntime.Must(imagev1.AddToScheme(scheme.Scheme))

	testEnv = testenv.New(testenv.WithCRDPath(filepath.Join("..", "..", "config", "crd", "bases")))

	var err error
	var badgerDir string
	badgerDir, err = os.MkdirTemp("", "badger")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temporary directory for badger: %v", err))
	}
	badgerOpts := badger.DefaultOptions(badgerDir)
	badgerOpts.Logger = nil
	testBadgerDB, err = badger.Open(badgerOpts)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new Badger database: %v", err))
	}

	if err = (&ImageRepositoryReconciler{
		Client:        testEnv,
		Database:      ircbadger.NewBadgerDatabase(testBadgerDB),
		EventRecorder: record.NewFakeRecorder(256),
	}).SetupWithManager(testEnv, ImageRepositoryReconcilerOptions{
		RateLimiter: controller.GetDefaultRateLimiter(),
	}); err != nil {
		panic(fmt.Sprintf("Failed to start ImageRepositoryReconciler: %v", err))
	}

	if err = (&ImagePolicyReconciler{
		Client:        testEnv,
		Database:      ircbadger.NewBadgerDatabase(testBadgerDB),
		EventRecorder: record.NewFakeRecorder(256),
	}).SetupWithManager(testEnv, ImagePolicyReconcilerOptions{
		RateLimiter: controller.GetDefaultRateLimiter(),
	}); err != nil {
		panic(fmt.Sprintf("Failed to start ImagePolicyReconciler: %v", err))
	}

	go func() {
		fmt.Println("Starting the test environment")
		if err := testEnv.Start(ctx); err != nil {
			panic(fmt.Sprintf("Failed to start the test environment manager: %v", err))
		}
	}()
	<-testEnv.Manager.Elected()

	code := m.Run()

	fmt.Println("Stopping the test environment")
	if err := testEnv.Stop(); err != nil {
		panic(fmt.Sprintf("Failed to stop the test environment: %v", err))
	}

	if err := testBadgerDB.Close(); err != nil {
		panic(fmt.Sprintf("Failed to close Badger: %v", err))
	}
	if err := os.RemoveAll(badgerDir); err != nil {
		panic(fmt.Sprintf("Failed to remove Badger dir: %v", err))
	}

	os.Exit(code)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
