package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Source: https://kafka.apache.org/protocol.html

type ClusterTopicPartition struct {
	ErrorCode                             int16 // 0 indicates NO_ERROR
	PartitionIndex                        int32
	LeaderID                              int32
	LeaderEpoch                           int32
	ReplicaNodeIDs                        []int32
	InsyncReplicaNodeIDs                  []int32
	EligibleLeaderReplicaNodeIDs          []int32
	LastKnownEligibleLeaderReplicaNodeIDs []int32
	OfflineReplicaNodeIDs                 []int32
}

type ClusterTopic struct {
	ErrorCode  int16
	TopicId    uuid.UUID
	Partitions []ClusterTopicPartition
	Batch      []byte
}

var clusterTopics map[string]*ClusterTopic = map[string]*ClusterTopic{}

func SetClusterTopics() error {
	time.Sleep(5 * time.Second) // allow the tester to write to the file
	data, err := os.ReadFile("/tmp/kraft-combined-logs/__cluster_metadata-0/00000000000000000000.log")
	if err != nil {
		return fmt.Errorf("unable to open log file, %w", err)
	}
	buf := bytes.NewBuffer(data)
	partitions := []struct {
		TopicID uuid.UUID
		ClusterTopicPartition
	}{}

	for {
		/*
			BatchOffset: (8 bytes)
			BatchLength: (4 bytes)
			PartitionLeaderEpoch: (4 bytes)
			MagicByte: (1 byte)
			RecordCRC32C: (4 bytes)
			Attribute: (2 bytes)
			LastOffSetDelta: (4 bytes)
			BaseTimestamp: (8 bytes)
			MaxTimestamp: (8 bytes)
			ProducerID: (8 bytes)
			ProducerEpoch: (2 bytes)
			BaseSequence: (4 bytes)
		*/
		startIndex := len(data) - buf.Len()
		_ = buf.Next(8 + 4 + 4 + 1 + 4 + 2 + 4 + 8 + 8 + 8 + 2 + 4)
		if buf.Len() == 0 {
			break
		}

		var recordLength int32
		_ = binary.Read(buf, binary.BigEndian, &recordLength)
		createdTopics := make([]string, 0)
		for range recordLength {
			size, _ := binary.ReadVarint(buf)
			recordBuf := bytes.NewBuffer(buf.Next(int(size)))
			_ = recordBuf.Next(1)                        // Attributes
			_, _ = binary.ReadVarint(recordBuf)          // Timestamp Delta
			_, _ = binary.ReadVarint(recordBuf)          // Offset Delta
			keyLength, _ := binary.ReadVarint(recordBuf) // Key Length
			if keyLength > 0 {
				_ = recordBuf.Next(int(keyLength))
			}
			valueLength, _ := binary.ReadVarint(recordBuf) // Value Length
			valueBuf := bytes.NewBuffer(recordBuf.Next(int(valueLength)))

			_ = valueBuf.Next(1) // Frame Version
			var recordType int8
			_ = binary.Read(valueBuf, binary.BigEndian, &recordType)
			_ = valueBuf.Next(1) // Version of feature-level record
			switch recordType {
			case 2: // TopicRecord
				topicNameLength, _ := binary.ReadUvarint(valueBuf)
				topicName, topicId := make([]byte, int(topicNameLength-1)), make([]byte, 16)
				_ = binary.Read(valueBuf, binary.BigEndian, &topicName)
				_ = binary.Read(valueBuf, binary.BigEndian, &topicId)

				id, _ := uuid.FromBytes(topicId)
				clusterTopics[string(topicName)] = &ClusterTopic{TopicId: id}
				createdTopics = append(createdTopics, string(topicName))
			case 3: // PartitionRecord
				var partition ClusterTopicPartition
				_ = binary.Read(valueBuf, binary.BigEndian, &partition.PartitionIndex)

				topicId := make([]byte, 16)
				_ = binary.Read(valueBuf, binary.BigEndian, &topicId)
				id, _ := uuid.FromBytes(topicId)

				replicaNodeIDsLength, _ := binary.ReadUvarint(valueBuf)
				partition.ReplicaNodeIDs = make([]int32, replicaNodeIDsLength)
				for i := range replicaNodeIDsLength {
					_ = binary.Read(valueBuf, binary.BigEndian, &partition.ReplicaNodeIDs[i])
				}

				insyncReplicaNodeIDsLength, _ := binary.ReadUvarint(valueBuf)
				partition.InsyncReplicaNodeIDs = make([]int32, insyncReplicaNodeIDsLength)
				for i := range insyncReplicaNodeIDsLength {
					_ = binary.Read(valueBuf, binary.BigEndian, &partition.InsyncReplicaNodeIDs[i])
				}

				_, _ = binary.ReadUvarint(valueBuf) // Removing Replicas Array Length
				_, _ = binary.ReadUvarint(valueBuf) // Adding Replicas Array Length
				_ = binary.Read(valueBuf, binary.BigEndian, &partition.LeaderID)
				_ = binary.Read(valueBuf, binary.BigEndian, &partition.LeaderEpoch)
				partitions = append(partitions, struct {
					TopicID uuid.UUID
					ClusterTopicPartition
				}{
					TopicID: id, ClusterTopicPartition: partition,
				})
			}
		}

		finalIndex := len(data) - buf.Len()
		for _, topic := range createdTopics {
			clusterTopics[topic].Batch = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0}, data[startIndex+8:finalIndex]...) // Reset baseOffset to 0
		}
	}

	for i := range partitions {
		for topicName := range clusterTopics {
			if clusterTopics[topicName].TopicId == partitions[i].TopicID {
				clusterTopics[topicName].Partitions = append(clusterTopics[topicName].Partitions, partitions[i].ClusterTopicPartition)
				break
			}
		}
	}

	return nil
}

func GetClusterTopic(topic string) *ClusterTopic {
	clusterTopic, ok := clusterTopics[topic]
	if ok {
		return clusterTopic
	}
	return &ClusterTopic{ErrorCode: 3}
}

type ApiKey struct {
	Key, MinVersion, MaxVersion int16
}

var (
	ApiKey_FetchRequest            ApiKey = ApiKey{Key: 1, MaxVersion: 16}
	ApiKey_ApiVersionRequest       ApiKey = ApiKey{Key: 18, MaxVersion: 4}
	ApiKey_DescribeTopicPartitions ApiKey = ApiKey{Key: 75}
)

type Request struct {
	ApiKey        int16
	ApiVersion    int16
	CorrelationId int32
	ClientId      []byte
	RequestBody   []byte
}

func ParseCompactArray(data []byte) (array []string, remainder []byte) {
	buf := bytes.NewBuffer(data)

	var length, size int8
	_ = binary.Read(buf, binary.BigEndian, &length)
	length--

	array = make([]string, length)
	for i := 0; i < int(length); i++ {
		_ = binary.Read(buf, binary.BigEndian, &size)
		size--

		value := make([]byte, size)
		_ = binary.Read(buf, binary.BigEndian, &value)
		array[i] = string(value)
		_ = buf.Next(1) // exclude tag_buffer
	}
	return array, buf.Bytes()
}

func ParseRequest(data []byte) *Request {
	/*
		Request:
			- message_size: position 0 - 3
			- header (v2)
				- request_api_key: position 4 - 5
				- request_api_version: position 6 - 7
				- correlation_id: position 8 - 11
				- client_id: (12 - 13 shows the length) variable in size
				- tag_buffer: one byte
			- body (dependent on the request)
	*/
	buf := bytes.NewBuffer(data)
	_ = buf.Next(4)

	var request Request
	_ = binary.Read(buf, binary.BigEndian, &request.ApiKey)
	_ = binary.Read(buf, binary.BigEndian, &request.ApiVersion)
	_ = binary.Read(buf, binary.BigEndian, &request.CorrelationId)

	var clientIdLength int16
	_ = binary.Read(buf, binary.BigEndian, &clientIdLength)
	if clientIdLength > 0 {
		request.ClientId = make([]byte, clientIdLength)
		_ = binary.Read(buf, binary.BigEndian, &request.ClientId)
	}

	_ = buf.Next(1) // exlude tag_buffer
	request.RequestBody = buf.Bytes()
	return &request
}

func ParseFetchRequest(data []byte) []uuid.UUID {
	/*
		MaxWaitMs: (4 bytes)
		MinBytes: (4 bytes)
		MaxBytes: (4 bytes)
		IsolationLevel: (1 byte)
		SessionID: (4 bytes)
		SessionEpoch: (4 bytes)
	*/
	buf := bytes.NewBuffer(data)
	_ = buf.Next(4 + 4 + 4 + 1 + 4 + 4)

	topicCount, _ := binary.ReadUvarint(buf)
	topicCount--

	topicIds := make([]uuid.UUID, 0, topicCount)
	for range topicCount {
		topicId := make([]byte, 16)
		_ = binary.Read(buf, binary.BigEndian, &topicId)
		id, _ := uuid.FromBytes(topicId)

		partitionCount, _ := binary.ReadUvarint(buf)
		partitionCount--

		for range partitionCount {
			/*
				PartitionID: (4 bytes)
				CurrentLeaderEpoch: (4 bytes)
				FetchOffset: (8 bytes)
				LastFetchedEpoch: (4 bytes)
				LogStartOffset: (8 bytes)
				PartitionMaxBytes: (4 bytes)
				TagBuffer: (1 byte)
			*/
			_ = buf.Next(4 + 4 + 8 + 4 + 8 + 4 + 1)
		}
		topicIds = append(topicIds, id)
	}
	return topicIds
}

func (r *Request) Fulfill() []byte {
	/*
		Response:
			- message_size: position 0 - 3
			- header (v0)
				- correlation_id: position 4 - 7
				- tag_buffer (if v1)
			- body (dependent on the request)
	*/
	b := new(bytes.Buffer)
	_ = binary.Write(b, binary.BigEndian, r.CorrelationId)

	switch r.ApiKey {
	case ApiKey_ApiVersionRequest.Key:
		var errorCode int16
		if r.ApiVersion < ApiKey_ApiVersionRequest.MinVersion || r.ApiVersion > ApiKey_ApiVersionRequest.MaxVersion {
			errorCode = 35
		}
		_ = binary.Write(b, binary.BigEndian, errorCode)
		_ = binary.Write(b, binary.BigEndian, int8(4)) // Number of api keys + 1

		_ = binary.Write(b, binary.BigEndian, ApiKey_ApiVersionRequest.Key)
		_ = binary.Write(b, binary.BigEndian, ApiKey_ApiVersionRequest.MinVersion)
		_ = binary.Write(b, binary.BigEndian, ApiKey_ApiVersionRequest.MaxVersion)
		_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER

		_ = binary.Write(b, binary.BigEndian, ApiKey_DescribeTopicPartitions.Key)
		_ = binary.Write(b, binary.BigEndian, ApiKey_DescribeTopicPartitions.MinVersion)
		_ = binary.Write(b, binary.BigEndian, ApiKey_DescribeTopicPartitions.MaxVersion)
		_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER

		_ = binary.Write(b, binary.BigEndian, ApiKey_FetchRequest.Key)
		_ = binary.Write(b, binary.BigEndian, ApiKey_FetchRequest.MinVersion)
		_ = binary.Write(b, binary.BigEndian, ApiKey_FetchRequest.MaxVersion)
		_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER

		_ = binary.Write(b, binary.BigEndian, int32(0)) // throttle_time_ms
		_ = binary.Write(b, binary.BigEndian, int8(0))  // TAG_BUFFER
	case ApiKey_DescribeTopicPartitions.Key:
		_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER

		_ = binary.Write(b, binary.BigEndian, int32(0)) // throttle_time_ms

		topicArray, _ := ParseCompactArray(r.RequestBody)
		_ = binary.Write(b, binary.BigEndian, int8(len(topicArray)+1))
		for _, topic := range topicArray {
			clusterTopic := GetClusterTopic(topic)

			_ = binary.Write(b, binary.BigEndian, clusterTopic.ErrorCode)               // error code meaning UNKNOWN_TOPIC_OR_PARTITION (3) or NO_ERROR (0)
			_ = binary.Write(b, binary.BigEndian, int8(len(topic)+1))                   // length of topic_name + 1
			_, _ = b.WriteString(topic)                                                 // topic_name
			_, _ = b.Write(clusterTopic.TopicId.Bytes())                                // topic_id
			_ = binary.Write(b, binary.BigEndian, false)                                // is_internal
			_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions)+1)) // length of partitions_array + 1
			for i := range clusterTopic.Partitions {
				_ = binary.Write(b, binary.BigEndian, clusterTopic.Partitions[i].ErrorCode)
				_ = binary.Write(b, binary.BigEndian, clusterTopic.Partitions[i].PartitionIndex)
				_ = binary.Write(b, binary.BigEndian, clusterTopic.Partitions[i].LeaderID)
				_ = binary.Write(b, binary.BigEndian, clusterTopic.Partitions[i].LeaderEpoch)
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions[i].ReplicaNodeIDs)+1))
				for _, id := range clusterTopic.Partitions[i].ReplicaNodeIDs {
					_ = binary.Write(b, binary.BigEndian, id)
				}
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions[i].InsyncReplicaNodeIDs)+1))
				for _, id := range clusterTopic.Partitions[i].InsyncReplicaNodeIDs {
					_ = binary.Write(b, binary.BigEndian, id)
				}
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions[i].EligibleLeaderReplicaNodeIDs)+1))
				for _, id := range clusterTopic.Partitions[i].EligibleLeaderReplicaNodeIDs {
					_ = binary.Write(b, binary.BigEndian, id)
				}
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions[i].LastKnownEligibleLeaderReplicaNodeIDs)+1))
				for _, id := range clusterTopic.Partitions[i].LastKnownEligibleLeaderReplicaNodeIDs {
					_ = binary.Write(b, binary.BigEndian, id)
				}
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Partitions[i].OfflineReplicaNodeIDs)+1))
				for _, id := range clusterTopic.Partitions[i].OfflineReplicaNodeIDs {
					_ = binary.Write(b, binary.BigEndian, id)
				}
				_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER
			}
			_ = binary.Write(b, binary.BigEndian, int32(0x00000df8)) // topic authorized questions
			_ = binary.Write(b, binary.BigEndian, int8(0))           // TAG_BUFFER
		}

		_ = binary.Write(b, binary.BigEndian, uint8(0xff)) // Next cursor; 0xff denotes a null value
		_ = binary.Write(b, binary.BigEndian, int8(0))     // TAG_BUFFER
	case ApiKey_FetchRequest.Key:
		_ = binary.Write(b, binary.BigEndian, int8(0))  // TAG_BUFFER
		_ = binary.Write(b, binary.BigEndian, int32(0)) // throttle_time_ms
		_ = binary.Write(b, binary.BigEndian, int16(0)) // error_code
		_ = binary.Write(b, binary.BigEndian, int32(0)) // session_id

		topicIds := ParseFetchRequest(r.RequestBody)
		_ = binary.Write(b, binary.BigEndian, int8(len(topicIds)+1)) // length of responses + 1
		for _, id := range topicIds {
			var clusterTopic *ClusterTopic
			for _, topic := range clusterTopics {
				if topic.TopicId == id {
					clusterTopic = topic
					break
				}
			}
			var errorCode int16
			if clusterTopic == nil {
				errorCode = 100
			}

			_ = binary.Write(b, binary.BigEndian, id.Bytes())
			_ = binary.Write(b, binary.BigEndian, int8(2)) // length of partitions + 1

			_ = binary.Write(b, binary.BigEndian, int32(0))  // partition_index
			_ = binary.Write(b, binary.BigEndian, errorCode) // error_code: 100; UNKNOWN_TOPIC or 0; NO_ERROR
			_ = binary.Write(b, binary.BigEndian, int64(0))  // high_watermark
			_ = binary.Write(b, binary.BigEndian, int64(0))  // last_stable_offset
			_ = binary.Write(b, binary.BigEndian, int64(0))  // log_start_offset
			_ = binary.Write(b, binary.BigEndian, int8(1))   // length of aborted_transactions + 1
			_ = binary.Write(b, binary.BigEndian, int32(0))  // preferred_read_replica
			if clusterTopic != nil && len(clusterTopic.Batch) > 0 {
				_ = binary.Write(b, binary.BigEndian, int8(len(clusterTopic.Batch))) // length of records batch
				_, _ = b.Write(clusterTopic.Batch)                                   // record batch
			} else {
				_ = binary.Write(b, binary.BigEndian, int8(0)) // length of records
			}
			_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER

			_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER
		}

		_ = binary.Write(b, binary.BigEndian, int8(0)) // TAG_BUFFER
	}
	messageSize := int32(len(b.Bytes()))

	response := new(bytes.Buffer)
	response.Grow(int(4 + messageSize))
	_ = binary.Write(response, binary.BigEndian, messageSize)
	_, _ = response.Write(b.Bytes())
	return response.Bytes()
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("unable to read from connection,", err.Error())
			return
		}
		request := ParseRequest(buf)
		_, _ = conn.Write(request.Fulfill())
		_ = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	}
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:9092")
	if err != nil {
		fmt.Println("Failed to bind to port 9092")
		os.Exit(1)
	}
	if err := SetClusterTopics(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("unable to accept a connection,", err.Error())
			break
		}
		wg.Add(1)
		_ = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
		go handleConnection(conn, &wg)
	}
	wg.Wait()
}
