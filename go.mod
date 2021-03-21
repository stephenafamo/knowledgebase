module github.com/stephenafamo/knowledgebase

go 1.15

require (
	github.com/go-chi/chi/v5 v5.0.1
	github.com/markbates/pkger v0.17.1
	github.com/mdigger/goldmark-images v0.0.0-20191226150935-49b26b7ee43c
	github.com/spf13/afero v1.4.0
	github.com/spf13/cobra v1.0.0
	github.com/stephenafamo/janus v0.0.0-20200917011258-52ecf63a4d75
	github.com/stephenafamo/orchestra v0.0.0-20200524112330-a21d225c0c33
	github.com/yuin/goldmark v1.2.1
)

replace github.com/stephenafamo/knowledgebase/examples => ./examples
