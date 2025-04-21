package client

import (
	"common/proto"
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	client := GetCosRpcClient(ctx)
	info, err := client.GetFileInfo(ctx, &proto.ResourceInfo{
		Fid: "fi_629a2743-d8db-41c5-be95-d8449f987343",
	})

	if err != nil {
		t.Error(err)
	}

	t.Log(info)
}
