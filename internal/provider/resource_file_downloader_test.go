// Copyright (c) Frontiers.gg
// SPDX-License-Identifier: MIT

package provider

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestFileResource_GET(t *testing.T) {
	want := []byte(testRandString(32))
	// Setup mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer ts.Close()

	sha1Sum := sha1.Sum(want)
	sha1Hex := hex.EncodeToString(sha1Sum[:])
	sha256Sum := sha256.Sum256(want)
	sha256Hex := hex.EncodeToString(sha256Sum[:])

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "utility_file_downloader" "file_test" {
						url = "%s"
						filename = "test_output.txt"
					}`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("utility_file_downloader.file_test", "id", sha1Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_test", "sha1", sha1Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_test", "sha256", sha256Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_test", "filename", "test_output.txt"),
					resource.TestCheckResourceAttrWith("utility_file_downloader.file_test", "filename", func(value string) error {
						got, err := os.ReadFile(value)
						if err != nil {
							return err
						}
						assert.Equal(t, want, got)
						return nil
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "utility_file_downloader" "file_test2" {
						url = "%s"
						method = "POST"
						filename = "test_output2.txt"

					}`, ts.URL),
				ExpectError: regexp.MustCompile(`failed to download file: 405 Method Not Allowed`),
			},
		},
	})
}

func TestFileResource_POST_WithHeaders(t *testing.T) {
	want := []byte(testRandString(32))
	// Setup mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Authorization") != "Bearer xyz" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer ts.Close()

	sha1Sum := sha1.Sum(want)
	sha1Hex := hex.EncodeToString(sha1Sum[:])
	sha256Sum := sha256.Sum256(want)
	sha256Hex := hex.EncodeToString(sha256Sum[:])

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "utility_file_downloader" "file_post_test" {
						url = "%s"
						method = "POST"
						filename = "test_post_output.txt"
						headers = {
							Authorization = "Bearer xyz"
						}
					}`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("utility_file_downloader.file_post_test", "id", sha1Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_post_test", "sha1", sha1Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_post_test", "sha256", sha256Hex),
					resource.TestCheckResourceAttr("utility_file_downloader.file_post_test", "filename", "test_post_output.txt"),
					resource.TestCheckResourceAttrWith("utility_file_downloader.file_post_test", "filename", func(value string) error {
						got, err := os.ReadFile(value)
						if err != nil {
							return err
						}
						assert.Equal(t, want, got)
						return nil
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "utility_file_downloader" "file_post_test2" {
						url = "%s"
						method = "POST"
						filename = "test_post_output2.txt"
						headers = {
							Authorization = "Bearer abc"
						}
					}`, ts.URL),
				ExpectError: regexp.MustCompile(`failed to download file: 401 Unauthorized`),
			},
		},
	})
}

func TestFileResource_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "utility_file_downloader" "file_failure" {
						url = "%s"
						filename = "failure.txt"
					}`, ts.URL),
				ExpectError: regexp.MustCompile(`failed to download file: 400 Bad Request`),
			},
		},
	})
}

var testLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func testRandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = testLetters[rand.Intn(len(testLetters))]
	}
	return string(b)
}
