server:
	hz update -idl ../idl/wcd.thrift
init_server:
	hz new -module github.com/DeepLangAI/wcd -idl ./wcd.thrift
manage:
	hz update -idl ./manage.thrift
dev_start:
	export MODE_ENV=dev && go run *.go

client_text_parse:
	hz client -idl ../idl/text_parse.thrift
