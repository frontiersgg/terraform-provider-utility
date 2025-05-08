

resource "utility_file_downloader" "example" {
  url      = "https://example.com/file.zip"
  filename = "${path.module}/file.zip"

  headers = {
    Authorization = "Bearer token"
  }
}
