package ctxz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bool64/ctxd"
	"github.com/bool64/logz/ctxz"
	"github.com/bool64/logz/logzpage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewObserver(t *testing.T) {
	logger := ctxd.NoOpLogger{}
	o := ctxz.NewObserver(logger)

	assert.Equal(t, o, o.CtxdLogger())

	ctx := ctxd.AddFields(context.Background(), "shared", 123)

	o.Debug(ctx, "debug", "foo", "bar")
	o.Info(ctx, "info", "foo", "bar")
	o.Important(ctx, "important", "foo", "bar")

	o = o.WithLogger(ctxd.NoOpLogger{})

	err := ctxd.NewError(ctx, "oops", "errDetail", 321)

	o.Warn(ctx, "warn", "foo", "bar")
	o.Error(ctx, "error", "foo", "bar", "error", err)

	req, err := http.NewRequest(http.MethodGet, "/debug/logz?level=Error&msg=error", nil)
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	logzpage.Handler(o.LevelObservers()...).ServeHTTP(rw, req)

	assert.Contains(t, rw.Body.String(), `{
 &#34;errDetail&#34;: 321,
 &#34;error&#34;: &#34;oops&#34;,
 &#34;foo&#34;: &#34;bar&#34;,
 &#34;shared&#34;: 123
}`)
}
