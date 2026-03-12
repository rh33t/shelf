package model

// This is the base directory for CTF challenges and boxes.
// Configure here before building or
// override with SHELF_BASE_DIR environment variable
var labDir = expandHome(envOrDefault("SHELF_BASE_DIR", "~/work"))

var ctfCategories = []string{
	"web-exploitation",
	"cryptography",
	"reverse-engineering",
	"forensics",
	"general",
	"binary-exploitation",
	"osint",
}

var boxPlatforms = []string{
	"tryhackme",
	"hackthebox",
}
