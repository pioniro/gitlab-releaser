package main

import (
	"io"
	"net/http"
	"reflect"
	"testing"
	"bytes"
)

func Test_check(t *testing.T) {
	type args struct {
		r *http.Request
	}
	config = &Config{}

	t.Run("check invalid token", func(t *testing.T) {
		req := &http.Request{Header:http.Header{}}
		req.Header.Set("X-Gitlab-Token", "wrong")
		config.Secret = "secret"
		if check(req) {
			t.Errorf("check() = %v, want %v", true, false)
		}
	})

	t.Run("check valid token", func(t *testing.T) {
		req := &http.Request{Header:http.Header{}}
		req.Header.Set("X-Gitlab-Token", "secret")
		config.Secret = "secret"
		if !check(req) {
			t.Errorf("check() = %v, want %v", false, true)
		}
	})

	t.Run("don't check token, when secret not set", func(t *testing.T) {
		req := &http.Request{Header:http.Header{}}
		req.Header.Set("X-Gitlab-Token", "secret")
		config.Secret = ""
		if !check(req) {
			t.Errorf("check() = %v, want %v", false, true)
		}
	})
}

func Test_decodeEvent(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want *PushEvent
	}{
		{
			name: "decode push event",
			args: args{
				r: bytes.NewBufferString(`
{
  "object_kind": "push",
  "event_name": "push",
  "before": "4c6a1fea4d79f56b41571d14c9ea967548875fb0",
  "after": "ad1794ea74a46c8bcb13755a793097c71330d92d",
  "ref": "refs/heads/master",
  "checkout_sha": "ad1794ea74a46c8bcb13755a793097c71330d92d",
  "message": null,
  "user_id": 3,
  "user_name": "Pioniro",
  "user_username": "Pioniro",
  "user_email": "pioniro@yandex.ru",
  "user_avatar": "https://gitlab.com/uploads/-/system/user/avatar/3/avatar.png",
  "project_id": 3,
  "project": {
    "id": 3,
    "name": "test",
    "description": "",
    "web_url": "https://gitlab.com/pioniro/test",
    "avatar_url": null,
    "git_ssh_url": "git@gitlab.com:pioniro/test.git",
    "git_http_url": "https://gitlab.com/pioniro/test.git",
    "namespace": "pioniro",
    "visibility_level": 0,
    "path_with_namespace": "pioniro/test",
    "default_branch": "master",
    "ci_config_path": null,
    "homepage": "https://gitlab.com/pioniro/test",
    "url": "git@gitlab.com:pioniro/test.git",
    "ssh_url": "git@gitlab.com:pioniro/test.git",
    "http_url": "https://gitlab.com/pioniro/test.git"
  },
  "commits": [
    {
      "id": "4c6a1fea4d79f56b41571d14c9ea967548875fb0",
      "message": "[Core] Fix comment count\n",
      "timestamp": "2018-11-14T10:08:15Z",
      "url": "https://gitlab.com/pioniro/test/commit/4c6a1fea4d79f56b41571d14c9ea967548875fb0",
      "author": {
        "name": "Aleksey Fedorov",
        "email": "pioniro@yandex.ru"
      }
    }
  ],
  "total_commits_count": 3,
  "repository": {
    "name": "test",
    "url": "git@gitlab.com:pioniro/test.git",
    "description": "",
    "homepage": "https://gitlab.com/pioniro/test"
  }
}
`),
			},
			want: &PushEvent{
				Sha: "ad1794ea74a46c8bcb13755a793097c71330d92d",
				Repository: &GitlabRepository{
					Homepage: "https://gitlab.com/pioniro/test",
				},
				Commits: []*GitlabCommit{
					{
						Id:        "4c6a1fea4d79f56b41571d14c9ea967548875fb0",
						Message:   "[Core] Fix comment count\n",
						Timestamp: "2018-11-14T10:08:15Z",
						Author: &PushCommitAuthor{
							Name:  "Aleksey Fedorov",
							Email: "pioniro@yandex.ru",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeEvent(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeEvent() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_buildPayloadFromPush(t *testing.T) {
	type args struct {
		event *PushEvent
	}
	tests := []struct {
		name string
		args args
		want *SentryPayload
	}{
		{
			name: "generate sentry request",
			args: args{
				event: &PushEvent{
					Sha: "ad1794ea74a46c8bcb13755a793097c71330d92d",
					Repository: &GitlabRepository{
						Homepage: "https://gitlab.com/pioniro/test",
					},
					Commits: []*GitlabCommit{
						{
							Id:        "4c6a1fea4d79f56b41571d14c9ea967548875fb0",
							Message:   "[Core] Fix comment count\n",
							Timestamp: "2018-11-14T10:08:15Z",
							Author: &PushCommitAuthor{
								Name:  "Aleksey Fedorov",
								Email: "pioniro@yandex.ru",
							},
						},
					},
				},
			},
			want: &SentryPayload{
				Version: "ad1794e",
				Ref:     "ad1794ea74a46c8bcb13755a793097c71330d92d",
				Url:     "https://gitlab.com/pioniro/test/tree/ad1794ea74a46c8bcb13755a793097c71330d92d",
				Commits: []*SentryPayloadCommit{
					{
						Id:          "4c6a1fea4d79f56b41571d14c9ea967548875fb0",
						Message:     "[Core] Fix comment count\n",
						AuthorName:  "Aleksey Fedorov",
						AuthorEmail: "pioniro@yandex.ru",
						Timestamp:   "2018-11-14T10:08:15Z",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildPayloadFromPush(tt.args.event); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildPayloadFromPush() = %v, want %v", got, tt.want)
			}
		})
	}
}
