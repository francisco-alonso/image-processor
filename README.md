# 📸 Cloud Image Processing Service

This project enables the processing of images uploaded to a **Google Cloud Storage bucket**.  
Whenever an image is uploaded to the **source bucket**, the service listens for a message in **Pub/Sub**, processes the image, and uploads it to a **destination bucket**.

---

## 🚀 Architecture

1. A user uploads an image to the **source bucket** in Google Cloud Storage.
2. **Cloud Pub/Sub** detects the upload and sends a message to a **topic**.
3. A **Cloud Run** service listens for messages from the topic, downloads the image, processes it (resizing), and uploads it to the **destination bucket**.

---

## 📌 1. Initial Setup

### 🔹 Enable Required APIs

```sh
gcloud services enable storage.googleapis.com
gcloud services enable pubsub.googleapis.com
gcloud services enable run.googleapis.com
```

---

## 📌 2. Creating Resources in Google Cloud

### 🔹 2.1. Define Environment Variables

```sh
export PROJECT_ID="your-project-id"
export REGION="us-central1"
export SOURCE_BUCKET="${PROJECT_ID}-source-images"
export DESTINATION_BUCKET="${PROJECT_ID}-destination-images"
export PUBSUB_TOPIC="image-processing-topic"
export SUBSCRIPTION_NAME="image-processing-sub"
export SERVICE_ACCOUNT="image-processor-sa"
```

### 🔹 2.2. Create the Cloud Storage Buckets

```sh
gcloud storage buckets create gs://$SOURCE_BUCKET --location=$REGION
gcloud storage buckets create gs://$DESTINATION_BUCKET --location=$REGION
```

### 🔹 2.3. Create the Pub/Sub Topic and Subscription

```sh
gcloud pubsub topics create $PUBSUB_TOPIC
gcloud pubsub subscriptions create $SUBSCRIPTION_NAME --topic=$PUBSUB_TOPIC
```

### 🔹 2.4. Configure Cloud Storage to Publish Events to Pub/Sub

```sh
gcloud storage buckets notifications create gs://$SOURCE_BUCKET \
  --topic=$PUBSUB_TOPIC \
  --event-types=OBJECT_FINALIZE \
  --message-format=json
```

---

## 📌 3. Configure the Service Account

### 🔹 3.1. Create the Service Account

```sh
gcloud iam service-accounts create $SERVICE_ACCOUNT \
  --display-name "Image Processor Service Account"
```

### 🔹 3.2. Assign Permissions to the Service Account

```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/storage.objectCreator"
```

### 🔹 3.3. Generate an Authentication Key for the Service Account for Local Testing

```sh
gcloud iam service-accounts keys create gcp-auth-key.json \
  --iam-account=${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com
```

---

## 📌 4. Build and Deploy to Cloud Run

### 🔹 4.1. Build and Push the Docker Image to Google Artifact Registry

```sh
gcloud builds submit --tag gcr.io/$PROJECT_ID/image-processor
```

### 🔹 4.2. Deploy the Service to Cloud Run

```sh
gcloud run deploy image-processor-service \
  --image gcr.io/$PROJECT_ID/image-processor \
  --platform managed \
  --region $REGION \
  --set-env-vars PROJECT_ID=$PROJECT_ID,SUBSCRIPTION_NAME=$SUBSCRIPTION_NAME,DESTINATION_BUCKET=$DESTINATION_BUCKET \
  --service-account=${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
  --allow-unauthenticated
```

---

## 📌 5. Testing and Validation

### 🔹 5.1. Upload an Image to the Source Bucket

```sh
gcloud storage cp test-image.jpg gs://$SOURCE_BUCKET/
```

### 🔹 5.2. Verify That the Processed Image Exists in the Destination Bucket

```sh
gcloud storage ls gs://$DESTINATION_BUCKET/
```

---

## 📌 6. Monitoring and Logs

### 🔹 View Logs in Cloud Run

```sh
gcloud run logs read image-processor-service
```

### 🔹 View Messages in Pub/Sub

```sh
gcloud pubsub subscriptions pull $SUBSCRIPTION_NAME --auto-ack
```

---

## 📌 7. Cleanup (Optional)

```sh
gcloud storage buckets delete gs://$SOURCE_BUCKET
gcloud storage buckets delete gs://$DESTINATION_BUCKET
gcloud pubsub topics delete $PUBSUB_TOPIC
gcloud pubsub subscriptions delete $SUBSCRIPTION_NAME
gcloud run services delete image-processor-service
gcloud iam service-accounts delete ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com
```

---

## 📌 8. Technical Explanation of the Code

The **Go service** performs the following actions:

1. **Listens for messages on Pub/Sub**: Extracts the bucket name and image file name from the JSON message.
2. **Downloads the image from the source bucket**.
3. **Processes the image** (resizing it to 300px wide).
4. **Uploads the processed image to the destination bucket**.
5. **Responds to HTTP requests** for local testing.

### 🔹 Dockerfile

The `Dockerfile` uses a **multi-stage build**:
- **Builder stage**: Compiles the Go binary inside a **Golang** image.
- **Final stage**: Uses a **Debian Slim** image, installs `ca-certificates`, and runs the binary.

```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o image-processor cmd/main.go

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /app/image-processor /app/
EXPOSE 8080
CMD ["/app/image-processor"]
```