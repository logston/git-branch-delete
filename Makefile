test/nodiff:
	rm -rf nodiff
	rm -rf git-branch-delete
	go build
	tar -xvzf testdata/nodiff.tgz
	cd nodiff && ../git-branch-delete
