// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package metaclient

import (
	"context"
	"errors"
	"strings"
	"time"

	"storj.io/common/encryption"
	"storj.io/common/paths"
	"storj.io/common/pb"
	"storj.io/common/storj"
)

var contentTypeKey = "content-type"

// Meta info about a segment.
type Meta struct {
	Modified   time.Time
	Expiration time.Time
	Size       int64
	Data       []byte
}

// GetObjectIPs returns the IP addresses of the nodes which hold the object.
func (db *DB) GetObjectIPs(ctx context.Context, bucket Bucket, key string) (_ *GetObjectIPsResponse, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket.Name == "" {
		return nil, ErrNoBucket.New("")
	}

	if key == "" {
		return nil, ErrNoPath.New("")
	}

	encPath, err := encryption.EncryptPathWithStoreCipher(bucket.Name, paths.NewUnencrypted(key), db.encStore)
	if err != nil {
		return nil, err
	}

	return db.metainfo.GetObjectIPs(ctx, GetObjectIPsParams{
		Bucket:        []byte(bucket.Name),
		EncryptedPath: []byte(encPath.Raw()),
	})
}

// CreateObject creates an uploading object and returns an interface for uploading Object information.
func (db *DB) CreateObject(ctx context.Context, bucket, key string, createInfo *CreateObject) (object *MutableObject, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return nil, ErrNoBucket.New("")
	}

	if key == "" {
		return nil, ErrNoPath.New("")
	}

	info := Object{
		Bucket: Bucket{Name: bucket},
		Path:   key,
	}

	if createInfo != nil {
		info.Metadata = createInfo.Metadata
		info.ContentType = createInfo.ContentType
		info.Expires = createInfo.Expires
		info.RedundancyScheme = createInfo.RedundancyScheme
		info.EncryptionParameters = createInfo.EncryptionParameters
	}

	// TODO: autodetect content type from the path extension
	// if info.ContentType == "" {}

	return &MutableObject{
		info: info,
	}, nil
}

// ModifyObject modifies a committed object.
func (db *DB) ModifyObject(ctx context.Context, bucket, key string) (object *MutableObject, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, errors.New("not implemented")
}

// DeleteObject deletes an object from database.
func (db *DB) DeleteObject(ctx context.Context, bucket, key string) (_ Object, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return Object{}, ErrNoBucket.New("")
	}

	if len(key) == 0 {
		return Object{}, ErrNoPath.New("")
	}

	encPath, err := encryption.EncryptPathWithStoreCipher(bucket, paths.NewUnencrypted(key), db.encStore)
	if err != nil {
		return Object{}, err
	}

	object, err := db.metainfo.BeginDeleteObject(ctx, BeginDeleteObjectParams{
		Bucket:        []byte(bucket),
		EncryptedPath: []byte(encPath.Raw()),
	})
	if err != nil {
		return Object{}, err
	}

	return db.objectFromRawObjectItem(ctx, bucket, key, object)
}

// ModifyPendingObject creates an interface for updating a partially uploaded object.
func (db *DB) ModifyPendingObject(ctx context.Context, bucket, key string) (object *MutableObject, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, errors.New("not implemented")
}

// ListPendingObjects lists pending objects in bucket based on the ListOptions.
func (db *DB) ListPendingObjects(ctx context.Context, bucket string, options ListOptions) (list ObjectList, err error) {
	defer mon.Task()(&ctx)(&err)
	return ObjectList{}, errors.New("not implemented")
}

// ListPendingObjectStreams lists streams for a specific pending object key.
func (db *DB) ListPendingObjectStreams(ctx context.Context, bucket string, options ListOptions) (list ObjectList, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return ObjectList{}, ErrNoBucket.New("")
	}

	var startAfter string
	switch options.Direction {
	// TODO for now we are supporting only After
	// case Forward:
	// 	// forward lists forwards from cursor, including cursor
	// 	startAfter = keyBefore(options.Cursor)
	case After:
		// after lists forwards from cursor, without cursor
		startAfter = options.Cursor
	default:
		return ObjectList{}, errClass.New("invalid direction %d", options.Direction)
	}

	pi, err := encryption.GetPrefixInfo(bucket, paths.NewUnencrypted(options.Prefix), db.encStore)
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	resp, err := db.metainfo.ListPendingObjectStreams(ctx, ListPendingObjectStreamsParams{
		Bucket:          []byte(bucket),
		EncryptedPath:   []byte(pi.PathEnc.Raw()),
		EncryptedCursor: []byte(startAfter),
		Limit:           int32(options.Limit),
	})
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	objectsList, err := db.pendingObjectsFromRawObjectList(ctx, resp.Items, pi, startAfter)
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	return ObjectList{
		Bucket: bucket,
		Prefix: options.Prefix,
		More:   resp.More,
		Items:  objectsList,
	}, nil
}

func (db *DB) pendingObjectsFromRawObjectList(ctx context.Context, items []RawObjectListItem, pi *encryption.PrefixInfo, startAfter string) (objectList []Object, err error) {
	objectList = make([]Object, 0, len(items))

	for _, item := range items {
		stream, streamMeta, err := TypedDecryptStreamInfo(ctx, pi.Bucket, pi.PathUnenc, item.EncryptedMetadata, db.encStore)
		if err != nil {
			// skip items that cannot be decrypted
			if encryption.ErrDecryptFailed.Has(err) {
				continue
			}
			return nil, errClass.Wrap(err)
		}

		object, err := db.objectFromRawObjectListItem(pi.Bucket, pi.PathUnenc.Raw(), item, stream, streamMeta)
		if err != nil {
			return nil, errClass.Wrap(err)
		}

		objectList = append(objectList, object)
	}

	return objectList, nil
}

// ListObjects lists objects in bucket based on the ListOptions.
func (db *DB) ListObjects(ctx context.Context, bucket string, options ListOptions) (list ObjectList, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return ObjectList{}, ErrNoBucket.New("")
	}

	if options.Prefix != "" && !strings.HasSuffix(options.Prefix, "/") {
		return ObjectList{}, errClass.New("prefix should end with slash")
	}

	var startAfter string
	switch options.Direction {
	// TODO for now we are supporting only After
	// case Forward:
	// 	// forward lists forwards from cursor, including cursor
	// 	startAfter = keyBefore(options.Cursor)
	case After:
		// after lists forwards from cursor, without cursor
		startAfter = options.Cursor
	default:
		return ObjectList{}, errClass.New("invalid direction %d", options.Direction)
	}

	// TODO: we should let libuplink users be able to determine what metadata fields they request as well
	// metaFlags := meta.All
	// if db.pathCipher(bucket) == EncNull || db.pathCipher(bucket) == EncNullBase64URL {
	// 	metaFlags = meta.None
	// }

	// TODO use flags with listing
	// if metaFlags&meta.Size != 0 {
	// Calculating the stream's size require also the user-defined metadata,
	// where stream store keeps info about the number of segments and their size.
	// metaFlags |= meta.UserDefined
	// }

	pi, err := encryption.GetPrefixInfo(bucket, paths.NewUnencrypted(options.Prefix), db.encStore)
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	startAfter, err = encryption.EncryptPathRaw(startAfter, pi.Cipher, &pi.ParentKey)
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	items, more, err := db.metainfo.ListObjects(ctx, ListObjectsParams{
		Bucket:          []byte(bucket),
		EncryptedPrefix: []byte(pi.ParentEnc.Raw()),
		EncryptedCursor: []byte(startAfter),
		Limit:           int32(options.Limit),
		Recursive:       options.Recursive,
		Status:          options.Status,
	})
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	objectsList, err := db.objectsFromRawObjectList(ctx, items, pi, startAfter)
	if err != nil {
		return ObjectList{}, errClass.Wrap(err)
	}

	return ObjectList{
		Bucket: bucket,
		Prefix: options.Prefix,
		More:   more,
		Items:  objectsList,
	}, nil
}

func (db *DB) objectsFromRawObjectList(ctx context.Context, items []RawObjectListItem, pi *encryption.PrefixInfo, startAfter string) (objectList []Object, err error) {
	objectList = make([]Object, 0, len(items))

	for _, item := range items {
		unencItem, err := encryption.DecryptPathRaw(string(item.EncryptedPath), pi.Cipher, &pi.ParentKey)
		if err != nil {
			// skip items that cannot be decrypted
			if encryption.ErrDecryptFailed.Has(err) {
				continue
			}
			return nil, errClass.Wrap(err)
		}

		var unencKey paths.Unencrypted
		if pi.ParentEnc.Valid() {
			unencKey = paths.NewUnencrypted(pi.ParentUnenc.Raw() + "/" + unencItem)
		} else {
			unencKey = paths.NewUnencrypted(unencItem)
		}

		stream, streamMeta, err := TypedDecryptStreamInfo(ctx, pi.Bucket, unencKey, item.EncryptedMetadata, db.encStore)
		if err != nil {
			// skip items that cannot be decrypted
			if encryption.ErrDecryptFailed.Has(err) {
				continue
			}
			return nil, errClass.Wrap(err)
		}

		object, err := db.objectFromRawObjectListItem(pi.Bucket, unencItem, item, stream, streamMeta)
		if err != nil {
			return nil, errClass.Wrap(err)
		}

		objectList = append(objectList, object)
	}

	return objectList, nil
}

// DownloadOptions contains additional options for downloading.
type DownloadOptions struct {
	Range StreamRange
}

// DownloadInfo contains response for DownloadObject.
type DownloadInfo struct {
	Object             storj.Object
	DownloadedSegments []DownloadSegmentWithRSResponse
	ListSegments       ListSegmentsResponse
	Range              StreamRange
}

// DownloadObject gets object information, lists segments and downloads the first segment.
func (db *DB) DownloadObject(ctx context.Context, bucket, key string, options DownloadOptions) (info DownloadInfo, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return DownloadInfo{}, storj.ErrNoBucket.New("")
	}
	if key == "" {
		return DownloadInfo{}, storj.ErrNoPath.New("")
	}

	encPath, err := encryption.EncryptPathWithStoreCipher(bucket, paths.NewUnencrypted(key), db.encStore)
	if err != nil {
		return DownloadInfo{}, err
	}

	resp, err := db.metainfo.DownloadObject(ctx, DownloadObjectParams{
		Bucket:             []byte(bucket),
		EncryptedObjectKey: []byte(encPath.Raw()),
		Range:              options.Range,
	})
	if err != nil {
		return DownloadInfo{}, err
	}

	return db.newDownloadInfo(ctx, bucket, key, resp, options.Range)
}

func (db *DB) newDownloadInfo(ctx context.Context, bucket, key string, response DownloadObjectResponse, streamRange StreamRange) (DownloadInfo, error) {
	object, err := db.objectFromRawObjectItem(ctx, bucket, key, response.Object)
	if err != nil {
		return DownloadInfo{}, err
	}

	return DownloadInfo{
		Object:             object,
		DownloadedSegments: response.DownloadedSegments,
		ListSegments:       response.ListSegments,
		Range:              streamRange.Normalize(object.Size),
	}, nil
}

// GetObject returns information about an object.
func (db *DB) GetObject(ctx context.Context, bucket, key string) (info Object, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return Object{}, ErrNoBucket.New("")
	}

	if key == "" {
		return Object{}, ErrNoPath.New("")
	}

	encPath, err := encryption.EncryptPathWithStoreCipher(bucket, paths.NewUnencrypted(key), db.encStore)
	if err != nil {
		return Object{}, err
	}

	objectInfo, err := db.metainfo.GetObject(ctx, GetObjectParams{
		Bucket:                     []byte(bucket),
		EncryptedPath:              []byte(encPath.Raw()),
		RedundancySchemePerSegment: true,
	})
	if err != nil {
		return Object{}, err
	}

	return db.objectFromRawObjectItem(ctx, bucket, key, objectInfo)
}

func (db *DB) objectFromRawObjectItem(ctx context.Context, bucket, key string, objectInfo RawObjectItem) (Object, error) {
	if objectInfo.Bucket == "" { // zero objectInfo
		return Object{}, nil
	}

	object := Object{
		Version:  0, // TODO:
		Bucket:   Bucket{Name: bucket},
		Path:     key,
		IsPrefix: false,

		Created:  objectInfo.Modified, // TODO: use correct field
		Modified: objectInfo.Modified, // TODO: use correct field
		Expires:  objectInfo.Expires,  // TODO: use correct field

		Stream: Stream{
			ID: objectInfo.StreamID,

			Size: objectInfo.PlainSize,

			RedundancyScheme:     objectInfo.RedundancyScheme,
			EncryptionParameters: objectInfo.EncryptionParameters,
		},
	}

	streamInfo, streamMeta, err := TypedDecryptStreamInfo(ctx, bucket, paths.NewUnencrypted(key), objectInfo.EncryptedMetadata, db.encStore)
	if err != nil {
		return Object{}, err
	}

	if object.Stream.EncryptionParameters.CipherSuite == storj.EncUnspecified {
		object.Stream.EncryptionParameters = storj.EncryptionParameters{
			CipherSuite: storj.CipherSuite(streamMeta.EncryptionType),
			BlockSize:   streamMeta.EncryptionBlockSize,
		}
	}
	if streamMeta.LastSegmentMeta != nil {
		var nonce storj.Nonce
		copy(nonce[:], streamMeta.LastSegmentMeta.KeyNonce)

		object.Stream.LastSegment = LastSegment{
			EncryptedKeyNonce: nonce,
			EncryptedKey:      streamMeta.LastSegmentMeta.EncryptedKey,
		}
	}

	err = updateObjectWithStream(&object, streamInfo, streamMeta)
	if err != nil {
		return Object{}, err
	}

	return object, nil
}

func (db *DB) objectFromRawObjectListItem(bucket string, path storj.Path, listItem RawObjectListItem, stream *pb.StreamInfo, streamMeta pb.StreamMeta) (Object, error) {
	object := Object{
		Version:  0, // TODO:
		Bucket:   Bucket{Name: bucket},
		Path:     path,
		IsPrefix: listItem.IsPrefix,

		Created:  listItem.CreatedAt, // TODO: use correct field
		Modified: listItem.CreatedAt, // TODO: use correct field
		Expires:  listItem.ExpiresAt,

		Stream: Stream{
			Size: listItem.PlainSize,
		},
	}

	object.Stream.ID = listItem.StreamID

	err := updateObjectWithStream(&object, stream, streamMeta)
	if err != nil {
		return Object{}, err
	}

	return object, nil
}

func updateObjectWithStream(object *Object, stream *pb.StreamInfo, streamMeta pb.StreamMeta) error {
	if stream == nil {
		return nil
	}

	serializableMeta := pb.SerializableMeta{}
	err := pb.Unmarshal(stream.Metadata, &serializableMeta)
	if err != nil {
		return err
	}

	// ensure that the map is not nil
	if serializableMeta.UserDefined == nil {
		serializableMeta.UserDefined = map[string]string{}
	}

	_, found := serializableMeta.UserDefined[contentTypeKey]
	if !found && serializableMeta.ContentType != "" {
		serializableMeta.UserDefined[contentTypeKey] = serializableMeta.ContentType
	}

	segmentCount := streamMeta.NumberOfSegments
	object.Metadata = serializableMeta.UserDefined

	if object.Stream.Size == 0 {
		object.Stream.Size = ((segmentCount - 1) * stream.SegmentsSize) + stream.LastSegmentSize
	}
	object.Stream.SegmentCount = segmentCount
	object.Stream.FixedSegmentSize = stream.SegmentsSize
	object.Stream.LastSegment.Size = stream.LastSegmentSize

	return nil
}

// MutableObject is for creating an object stream.
type MutableObject struct {
	info Object
}

// Info gets the current information about the object.
func (object *MutableObject) Info() Object { return object.info }

// CreateStream creates a new stream for the object.
func (object *MutableObject) CreateStream(ctx context.Context) (_ *MutableStream, err error) {
	defer mon.Task()(&ctx)(&err)
	return &MutableStream{
		info: object.info,
	}, nil
}

// CreateDynamicStream creates a new dynamic stream for the object.
func (object *MutableObject) CreateDynamicStream(ctx context.Context, metadata SerializableMeta, expires time.Time) (_ *MutableStream, err error) {
	defer mon.Task()(&ctx)(&err)
	return &MutableStream{
		info: object.info,

		dynamic:         true,
		dynamicMetadata: metadata,
		dynamicExpires:  expires,
	}, nil
}

// TypedDecryptStreamInfo decrypts stream info.
func TypedDecryptStreamInfo(ctx context.Context, bucket string, unencryptedKey paths.Unencrypted, streamMetaBytes []byte, encStore *encryption.Store) (
	_ *pb.StreamInfo, streamMeta pb.StreamMeta, err error) {
	defer mon.Task()(&ctx)(&err)

	err = pb.Unmarshal(streamMetaBytes, &streamMeta)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	if encStore.EncryptionBypass {
		return nil, streamMeta, nil
	}

	derivedKey, err := encryption.DeriveContentKey(bucket, unencryptedKey, encStore)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	cipher := storj.CipherSuite(streamMeta.EncryptionType)
	encryptedKey, keyNonce := getEncryptedKeyAndNonce(streamMeta.LastSegmentMeta)
	contentKey, err := encryption.DecryptKey(encryptedKey, cipher, derivedKey, keyNonce)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	// decrypt metadata with the content encryption key and zero nonce
	streamInfo, err := encryption.Decrypt(streamMeta.EncryptedStreamInfo, cipher, contentKey, &storj.Nonce{})
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	var stream pb.StreamInfo
	if err := pb.Unmarshal(streamInfo, &stream); err != nil {
		return nil, pb.StreamMeta{}, err
	}

	return &stream, streamMeta, nil
}

func getEncryptedKeyAndNonce(m *pb.SegmentMeta) (storj.EncryptedPrivateKey, *storj.Nonce) {
	if m == nil {
		return nil, nil
	}

	var nonce storj.Nonce
	copy(nonce[:], m.KeyNonce)

	return m.EncryptedKey, &nonce
}
