package service

import (
	"application/pkg/fabric" // Assuming fabric interaction logic is here
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings" // Import strings package
	"time"

	"github.com/google/uuid" // For generating unique CertID and filenames
)

// CertificateService handles certificate related operations
type CertificateService struct{}

// Certificate structure matching the chaincode one (for JSON marshaling/unmarshaling)
type CertificatePayload struct {
	CertID       string    `json:"certId"`
	CarID        string    `json:"carId"`
	CertType     string    `json:"certType"`
	FileHash     string    `json:"fileHash"`
	FileLocation string    `json:"fileLocation"`
	UploadTime   time.Time `json:"uploadTime"`
}

const certificateBaseDir = "data/certificates" // Base directory for storing cert files relative to server root

// AddCertificate handles saving the file, calculating hash, and calling chaincode
func (s *CertificateService) AddCertificate(carId string, certType string, fileHeader *multipart.FileHeader) (*CertificatePayload, error) {
	// --- Check if certificate already exists for this car ---
	existingCerts, err := s.GetCertificatesByCar(carId) // First declaration of err in this function
	if err != nil {
		// Don't block upload if the check itself fails, but log it.
		fmt.Printf("警告: 检查车辆 %s 的现有证书时出错: %v\n", carId, err)
	} else if len(existingCerts) > 0 {
		return nil, fmt.Errorf("车辆 %s 已存在证书，无法重复上传", carId)
	}
	// --- End check ---

	// 1. Generate unique ID and determine file path
	certID := uuid.New().String()
	fileExt := filepath.Ext(fileHeader.Filename)
	fileName := certID + fileExt // Use CertID as filename to ensure uniqueness
	carCertDir := filepath.Join(certificateBaseDir, carId)
	filePath := filepath.Join(carCertDir, fileName)
	fileLocation := filepath.ToSlash(filePath) // Use forward slashes for consistency

	// 2. Ensure directory exists
	err = os.MkdirAll(carCertDir, os.ModePerm) // Assign to existing err
	if err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 3. Open the uploaded file
	srcFile, err := fileHeader.Open() // Assign to existing err, declare srcFile
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer srcFile.Close()

	// 4. Create the destination file
	dstFile, err := os.Create(filePath) // Assign to existing err, declare dstFile
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	// Use defer to ensure dstFile is closed eventually
	defer func() {
		if cerr := dstFile.Close(); cerr != nil && err == nil { // Only assign to err if no previous error occurred
			err = fmt.Errorf("关闭目标文件 %s 失败: %v", filePath, cerr)
			// If closing fails, we might still want to try removing the file
			os.Remove(filePath)
		}
	}()

	// 5. Copy content and calculate hash simultaneously
	hasher := sha256.New()
	// Use TeeReader to write to file and hasher at the same time
	_, err = io.Copy(dstFile, io.TeeReader(srcFile, hasher)) // Assign to existing err
	if err != nil {
		// Attempt to remove the potentially partially written file
		// defer dstFile.Close() will still run
		os.Remove(filePath)
		return nil, fmt.Errorf("保存文件并计算哈希失败: %v", err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Explicitly close dstFile *before* chaincode call to ensure data is flushed
	// The defer will still run but should be harmless on an already closed file.
	if cerr := dstFile.Close(); cerr != nil {
		os.Remove(filePath) // Attempt removal if close fails here too
		return nil, fmt.Errorf("写入后关闭目标文件 %s 失败: %v", filePath, cerr)
	}

	// 6. Prepare payload for chaincode
	uploadTime := time.Now()
	certPayload := CertificatePayload{
		CertID:       certID,
		CarID:        carId,
		CertType:     certType,
		FileHash:     fileHash,
		FileLocation: fileLocation, // Store relative path
		UploadTime:   uploadTime,
	}
	certJsonBytes, err := json.Marshal(certPayload) // Assign to existing err
	if err != nil {
		// Attempt to remove the saved file if marshaling fails
		os.Remove(filePath)
		return nil, fmt.Errorf("序列化证书数据失败: %v", err)
	}

	// 7. Call chaincode (Assuming org1 submits certificate for now, adjust if needed)
	contract := fabric.GetContract(CAR_DEALER_ORG)                               // Or use a dedicated org/identity if applicable
	_, err = contract.SubmitTransaction("AddCertificate", string(certJsonBytes)) // Assign to existing err
	if err != nil {
		// Attempt to remove the saved file if chaincode submission fails
		os.Remove(filePath)
		return nil, fmt.Errorf("调用链码 AddCertificate 失败: %s", fabric.ExtractErrorMessage(err))
	}

	// If everything succeeded up to here, err should be nil
	return &certPayload, err // Return the final err state (should be nil on success)
}

// GetCertificatesByCar retrieves all certificates for a car by filtering results from chaincode
func (s *CertificateService) GetCertificatesByCar(carId string) ([]*CertificatePayload, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG) // Or use a dedicated org/identity if applicable

	resultBytes, err := contract.EvaluateTransaction("GetAllCertificates") // First declaration of err in this function
	if err != nil {
		errMsg := fabric.ExtractErrorMessage(err)
		if strings.Contains(errMsg, "不存在") || strings.Contains(strings.ToLower(errMsg), "not found") {
			return []*CertificatePayload{}, nil // Return empty list if no certs found
		}
		return nil, fmt.Errorf("调用链码 GetAllCertificates 失败: %s", errMsg)
	}

	if len(resultBytes) == 0 || string(resultBytes) == "null" || string(resultBytes) == "[]" {
		return []*CertificatePayload{}, nil
	}

	var allCertificates []*CertificatePayload
	err = json.Unmarshal(resultBytes, &allCertificates) // Assign to existing err
	if err != nil {
		return nil, fmt.Errorf("解析链码返回的证书列表失败: %v, Raw: %s", err, string(resultBytes))
	}

	var carCertificates []*CertificatePayload
	for _, cert := range allCertificates {
		if cert != nil && cert.CarID == carId {
			carCertificates = append(carCertificates, cert)
		}
	}

	return carCertificates, nil
}

// GetCertificateFileLocation retrieves the file path for a given certificate ID
func (s *CertificateService) GetCertificateFileLocation(certId string) (string, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)

	allCertsBytes, err := contract.EvaluateTransaction("GetAllCertificates") // First declaration of err in this function
	if err != nil {
		errMsg := fabric.ExtractErrorMessage(err)
		if strings.Contains(errMsg, "不存在") || strings.Contains(strings.ToLower(errMsg), "not found") {
			return "", fmt.Errorf("证书 %s 未找到 (no certs on chain)", certId)
		}
		return "", fmt.Errorf("调用链码 GetAllCertificates 失败: %s", errMsg)
	}

	if len(allCertsBytes) == 0 || string(allCertsBytes) == "null" || string(allCertsBytes) == "[]" {
		return "", fmt.Errorf("证书 %s 未找到 (empty result from chain)", certId)
	}

	var allCertificates []*CertificatePayload
	err = json.Unmarshal(allCertsBytes, &allCertificates) // Assign to existing err
	if err != nil {
		return "", fmt.Errorf("解析链码返回的证书列表失败: %v, Raw: %s", err, string(allCertsBytes))
	}

	for _, cert := range allCertificates {
		if cert != nil && cert.CertID == certId {
			absPath, err := filepath.Abs(cert.FileLocation) // Shadow err is okay here
			if err != nil {
				fmt.Printf("警告: 无法获取绝对路径 for %s: %v\n", cert.FileLocation, err)
				if filepath.IsAbs(cert.FileLocation) {
					return cert.FileLocation, nil
				}
				return cert.FileLocation, nil
			}
			return absPath, nil
		}
	}

	return "", fmt.Errorf("证书 %s 未找到 (not in list)", certId)
}

// VerifyCertificate compares the stored hash from the blockchain with the hash of the actual file on the server.
func (s *CertificateService) VerifyCertificate(certId string) (bool, string, string, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)
	resultBytes, err := contract.EvaluateTransaction("GetCertificate", certId) // First declaration of err in this function
	if err != nil {
		errMsg := fabric.ExtractErrorMessage(err)
		if strings.Contains(errMsg, "不存在") || strings.Contains(strings.ToLower(errMsg), "not found") {
			return false, "", "", fmt.Errorf("证书 %s 在链上未找到", certId)
		}
		return false, "", "", fmt.Errorf("调用链码 GetCertificate(%s) 失败: %s", certId, errMsg)
	}
	if len(resultBytes) == 0 || string(resultBytes) == "null" {
		return false, "", "", fmt.Errorf("证书 %s 在链上未找到 (empty result)", certId)
	}

	var certPayload CertificatePayload
	err = json.Unmarshal(resultBytes, &certPayload) // Assign to existing err
	if err != nil {
		return false, "", "", fmt.Errorf("解析链码返回的证书数据失败: %v, Raw: %s", err, string(resultBytes))
	}

	storedHash := certPayload.FileHash
	fileLocation := certPayload.FileLocation

	if storedHash == "" || fileLocation == "" {
		return false, "", "", fmt.Errorf("链上证书记录缺少哈希或文件路径信息")
	}

	filePath := filepath.Join(".", fileLocation)

	if _, err = os.Stat(filePath); os.IsNotExist(err) { // Assign to existing err
		return false, storedHash, "", fmt.Errorf("服务器上找不到文件: %s", filePath)
	}

	file, err := os.Open(filePath) // Assign to existing err, declare file
	if err != nil {
		return false, storedHash, "", fmt.Errorf("打开服务器文件 %s 失败: %v", filePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file) // Assign to existing err
	if err != nil {
		return false, storedHash, "", fmt.Errorf("计算文件 %s 哈希失败: %v", filePath, err)
	}
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	match := storedHash == currentHash

	return match, storedHash, currentHash, nil
}

// VerifyUploadedCertificate compares the hash of an uploaded file with the hash of the original certificate stored on the blockchain for a given car.
func (s *CertificateService) VerifyUploadedCertificate(carId string, fileHeader *multipart.FileHeader) (bool, string, string, error) {
	originalCerts, err := s.GetCertificatesByCar(carId) // First declaration of err in this function
	if err != nil {
		return false, "", "", fmt.Errorf("获取车辆 %s 的原始证书信息失败: %v", carId, err)
	}
	if len(originalCerts) == 0 {
		return false, "", "", fmt.Errorf("车辆 %s 没有已上传的原始证书记录，无法进行比对", carId)
	}
	originalCert := originalCerts[0]
	storedHash := originalCert.FileHash

	if storedHash == "" {
		return false, "", "", fmt.Errorf("链上原始证书记录缺少哈希信息")
	}

	uploadedFile, err := fileHeader.Open() // Assign to existing err, declare uploadedFile
	if err != nil {
		return false, storedHash, "", fmt.Errorf("打开待验证文件失败: %v", err)
	}
	defer uploadedFile.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, uploadedFile) // Assign to existing err
	if err != nil {
		return false, storedHash, "", fmt.Errorf("计算待验证文件哈希失败: %v", err)
	}
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	match := storedHash == currentHash

	return match, storedHash, currentHash, nil
}
