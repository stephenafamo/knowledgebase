module github.com/stephenafamo/knowledgebase

go 1.15

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/markbates/pkger v0.17.1
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/mdigger/goldmark-images v0.0.0-20191226150935-49b26b7ee43c
	github.com/spf13/afero v1.4.0
	github.com/spf13/cobra v1.0.0
	github.com/stephenafamo/janus v0.0.0-20200917011258-52ecf63a4d75
	github.com/stephenafamo/knowledgebase/examples v0.0.0-00010101000000-000000000000 // indirect
	github.com/stephenafamo/orchestra v0.0.0-20200524112330-a21d225c0c33
	github.com/yuin/goldmark v1.2.1
)

replace github.com/go-chi/chi => github.com/stephenafamo/chi v4.2.0+incompatible

replace github.com/stephenafamo/knowledgebase/examples => ./examples
