package infrastructure

import (
	"context"

	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/yhartanto178dev/api-archiven-v2/domain"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/configs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoFileRepository struct {
	// client *mongo.Client
	db *mongo.Database
}

func NewMongoFileRepository(db *mongo.Database) *MongoFileRepository {
	return &MongoFileRepository{db: db}
}

func (r *MongoFileRepository) gridFSBucket() (*gridfs.Bucket, error) {
	bucket, err := gridfs.NewBucket(r.db, options.GridFSBucket().SetName("files"))
	return bucket, errors.Wrap(err, "failed to create GridFS bucket")
}

func (r *MongoFileRepository) Save(file *domain.File, content io.Reader) error {
	bucket, err := r.gridFSBucket()
	if err != nil {
		return err
	}

	uploadOpts := options.GridFSUpload().
		SetMetadata(bson.M{"contentType": file.ContentType})

	uploadStream, err := bucket.OpenUploadStream(file.Name, uploadOpts)
	if err != nil {
		return errors.Wrap(err, "failed to open upload stream")
	}
	defer uploadStream.Close()

	size, err := io.Copy(uploadStream, content)
	if err != nil {
		configs.Logger.Errorw("failed to save file to GridFS",
			"error", err.Error(),
			"file_name", file.Name,
			"operation", "save",
		)
		return errors.Wrap(err, "failed to write content to GridFS")
	}

	configs.Logger.Infow("file saved successfully",
		"file_id", file.ID,
		"file_size", file.Size,
	)

	file.ID = uploadStream.FileID.(primitive.ObjectID).Hex()
	file.Size = size
	return nil
}

func (r *MongoFileRepository) FindByID(id string) (*domain.File, io.ReadCloser, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil, domain.ErrFileNotFound
	}

	bucket, err := r.gridFSBucket()
	if err != nil {
		return nil, nil, err
	}

	downloadStream, err := bucket.OpenDownloadStream(objID)
	if err != nil {
		if err == gridfs.ErrFileNotFound {
			return nil, nil, domain.ErrFileNotFound
		}
		return nil, nil, errors.Wrap(err, "failed to open download stream")
	}

	fileDoc := downloadStream.GetFile()
	file := &domain.File{
		ID:          fileDoc.ID.(primitive.ObjectID).Hex(),
		Name:        fileDoc.Name,
		Size:        fileDoc.Length,
		ContentType: fileDoc.Metadata.Lookup("contentType").StringValue(),
		UploadDate:  fileDoc.UploadDate,
	}

	return file, downloadStream, nil
}

func (r *MongoFileRepository) FindAll(skip, limit int64) ([]*domain.File, error) {
	// Add pagination to the query
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.D{{Key: "uploadDate", Value: -1}}) // Sort by newest first

	// cursor, err := r.db.Collection("fs.files").Find(
	// 	context.Background(),
	// 	bson.M{},
	// 	findOptions,
	// )
	bucket, err := r.gridFSBucket()
	if err != nil {
		return nil, err
	}

	cursor, err := bucket.GetFilesCollection().Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find files")
	}
	defer cursor.Close(context.Background())

	var files []*domain.File
	for cursor.Next(context.Background()) {
		var fileDoc struct {
			ID         primitive.ObjectID `bson:"_id"`
			Name       string             `bson:"filename"`
			Length     int64              `bson:"length"`
			UploadDate time.Time          `bson:"uploadDate"`
			Metadata   bson.M             `bson:"metadata"`
		}

		if err := cursor.Decode(&fileDoc); err != nil {
			return nil, errors.Wrap(err, "failed to decode file document")
		}

		contentType := ""
		if fileDoc.Metadata != nil {
			contentType, _ = fileDoc.Metadata["contentType"].(string)
		}

		files = append(files, &domain.File{
			ID:          fileDoc.ID.Hex(),
			Name:        fileDoc.Name,
			Size:        fileDoc.Length,
			ContentType: contentType,
			UploadDate:  fileDoc.UploadDate,
		})
	}

	return files, nil
}

func (r *MongoFileRepository) Count() (int64, error) {
	bucket, err := r.gridFSBucket()
	if err != nil {
		return 0, err
	}
	cursor, _ := bucket.GetFilesCollection().CountDocuments(context.Background(), bson.M{})
	return cursor, nil
}
