build:
	go build

test/diff: build
	tar -xvzf testdata/diff.tgz
	cd diff && ../git-branch-delete

test/nodiff: build
	tar -xvzf testdata/nodiff.tgz
	cd nodiff && ../git-branch-delete

test/rebasefail: build
	tar -xvzf testdata/rebasefail.tgz
	cd rebasefail && ../git-branch-delete

test/all: clean test/diff test/nodiff test/rebasefail

clean:
	rm -rf git-branch-delete
	rm -rf diff
	rm -rf nodiff
	rm -rf rebasefail
