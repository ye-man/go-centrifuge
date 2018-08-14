package grpcservice

import (
	"context"
	"fmt"
	"time"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes"
)

func incompatibleVersionError(nodeVersion string) error {
	return errors.New(errors.VersionMismatch, fmt.Sprintf("Incompatible version: node version: %s, client version: %s", version.GetVersion(), nodeVersion))
}

func incompatibleNetworkError(nodeNetwork uint32) error {
	return errors.New(errors.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", config.Config.GetNetworkID(), nodeNetwork))
}

type P2PService struct {
	Notifier notification.Sender
}

// HandleP2PPost does the basic P2P handshake, stores the document received and sends notification to listener.
// It currently does not do any more processing.
//
// The handshake is currently quite primitive as it only allows the request-server
// to recipient to determine if two versions are compatible. A newer node making a
// request could not decide for itself if the request handshake should succeed or not.
func (srv *P2PService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	// Check call compatibility:
	compatible, err := version.CheckMajorCompatibility(req.CentNodeVersion)
	if err != nil {
		return nil, errors.New(errors.Unknown, err.Error())
	}
	if !compatible {
		return nil, incompatibleVersionError(req.CentNodeVersion)
	}

	if req.NetworkIdentifier != config.Config.GetNetworkID() {
		return nil, incompatibleNetworkError(req.NetworkIdentifier)
	}

	if req.Document == nil {
		return nil, errors.New(errors.DocumentInvalid, errors.NilError(req.Document).Error())
	}

	err = coredocumentrepository.GetCoreDocumentRepository().CreateOrUpdate(req.Document)
	if err != nil {
		return nil, errors.New(errors.Unknown, err.Error())
	}

	ts, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, errors.New(errors.Unknown, err.Error())
	}

	notificationMsg := &notificationpb.NotificationMessage{
		EventType:    uint32(notification.RECEIVED_PAYLOAD),
		CentrifugeId: req.SenderCentrifugeId,
		Recorded:     ts,
		Document:     req.Document,
	}

	// Async until we add queuing
	go srv.Notifier.Send(notificationMsg)

	return &p2ppb.P2PReply{
		CentNodeVersion: version.GetVersion().String(),
		Document:        req.Document,
	}, nil
}
