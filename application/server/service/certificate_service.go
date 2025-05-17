package service

import (
	"application/pkg/fabric" // 假设fabric交互逻辑在这里
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings" // 导入strings包
	"time"

	"github.com/google/uuid" // 用于生成唯一的CertID和文件名
)

// CertificateService 处理证书相关操作
type CertificateService struct{}

// Certificate 结构体匹配链码中的结构（用于JSON序列化/反序列化）
type CertificatePayload struct {
	CertID       string    `json:"certId"`
	CarID        string    `json:"carId"`
	CertType     string    `json:"certType"`
	FileHash     string    `json:"fileHash"`
	FileLocation string    `json:"fileLocation"`
	UploadTime   time.Time `json:"uploadTime"`
}

const certificateBaseDir = "data/certificates" // 相对于服务器根目录存储证书文件的基础目录

// AddCertificate 处理保存文件、计算哈希并调用链码
func (s *CertificateService) AddCertificate(carId string, certType string, fileHeader *multipart.FileHeader) (*CertificatePayload, error) {
	// --- 检查该车辆是否已存在证书 ---
	existingCerts, err := s.GetCertificatesByCar(carId) // 此函数中err的首次声明
	if err != nil {
		// 如果检查本身失败，不要阻止上传，但要记录它
		fmt.Printf("警告: 检查车辆 %s 的现有证书时出错: %v\n", carId, err)
	} else if len(existingCerts) > 0 {
		return nil, fmt.Errorf("车辆 %s 已存在证书，无法重复上传", carId)
	}
	// --- 检查结束 ---

	// 1. 生成唯一ID并确定文件路径
	certID := uuid.New().String()
	fileExt := filepath.Ext(fileHeader.Filename)
	fileName := certID + fileExt // 使用CertID作为文件名以确保唯一性
	carCertDir := filepath.Join(certificateBaseDir, carId)
	// serverFilePath是服务器文件系统上的实际路径
	serverFilePath := filepath.Join(carCertDir, fileName)
	// chaincodeFileLocation是存储在链码中的路径，相对于certificateBaseDir并以CarID开头
	chaincodeFileLocation := filepath.ToSlash(filepath.Join(carId, fileName))

	// 2. 确保目录存在
	err = os.MkdirAll(carCertDir, os.ModePerm) // 赋值给现有的err
	if err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 3. 打开上传的文件
	srcFile, err := fileHeader.Open() // 赋值给现有的err，声明srcFile
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer srcFile.Close()

	// 4. 创建目标文件
	dstFile, err := os.Create(serverFilePath) // 使用serverFilePath
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	fmt.Printf("DEBUG: 文件尝试创建于 %s\n", serverFilePath) // DEBUG LOG 1

	// 使用defer确保dstFile最终被关闭
	// 这个defer主要用于函数因错误退出或显式关闭本身出现问题的情况
	defer func() {
		fmt.Printf("DEBUG: Defer close for %s is executing. Outer err: %v\n", serverFilePath, err)
		if cerr := dstFile.Close(); cerr != nil {
			fmt.Printf("DEBUG: Defer close for %s 失败: %v\n", serverFilePath, cerr)
			// 仅当外部err尚未被前面的错误设置时才设置外部err
			if err == nil {
				err = fmt.Errorf("在defer中关闭目标文件 %s 失败: %v", serverFilePath, cerr)
				// 考虑如果defer关闭失败且没有其他错误，是否需要os.Remove
				// 目前，专注于记录
			}
		} else {
			fmt.Printf("DEBUG: Defer close for %s 成功.\n", serverFilePath)
		}
	}()

	// 5. 同时复制内容并计算哈希
	hasher := sha256.New()
	// 使用TeeReader同时写入文件和哈希器
	_, err = io.Copy(dstFile, io.TeeReader(srcFile, hasher)) // 赋值给现有的err
	if err != nil {
		fmt.Printf("DEBUG: io.Copy 过程中发生错误: %v (路径: %s)。尝试删除文件。\n", err, serverFilePath)
		os.Remove(serverFilePath)
		return nil, fmt.Errorf("保存文件并计算哈希失败: %v", err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("DEBUG: 文件内容已复制且哈希已计算 %s\n", serverFilePath) // DEBUG LOG 2

	// 在调用链码之前显式关闭dstFile，确保数据已刷新
	// defer仍会运行，但如果由操作系统处理，对已关闭的文件应该无害
	if cerr := dstFile.Close(); cerr != nil {
		fmt.Printf("DEBUG: 显式 dstFile.Close() 过程中发生错误: %v (路径: %s)。尝试删除文件。\n", cerr, serverFilePath)
		os.Remove(serverFilePath)
		return nil, fmt.Errorf("写入后关闭目标文件 %s 失败: %v", serverFilePath, cerr)
	}
	fmt.Printf("DEBUG: 文件已显式关闭，调用链码前应存在于 %s。\n", serverFilePath) // DEBUG LOG 3

	// 6. 准备链码的payload
	uploadTime := time.Now()
	certPayload := CertificatePayload{
		CertID:       certID,
		CarID:        carId,
		CertType:     certType,
		FileHash:     fileHash,
		FileLocation: chaincodeFileLocation, // 存储链码特定的相对路径
		UploadTime:   uploadTime,
	}
	certJsonBytes, err := json.Marshal(certPayload) // 赋值给现有的err
	if err != nil {
		fmt.Printf("DEBUG: json.Marshal 过程中发生错误: %v。尝试删除文件 %s。\n", err, serverFilePath)
		os.Remove(serverFilePath)
		return nil, fmt.Errorf("序列化证书数据失败: %v", err)
	}

	// 7. 调用链码（假设org1现在提交证书，如果需要可以调整）
	fmt.Printf("DEBUG: 即将使用以下payload调用 AddCertificate 的 SubmitTransaction: %s\n", string(certJsonBytes)) // DEBUG LOG 4
	contract := fabric.GetContract(CAR_DEALER_ORG)                                                       // 或使用专用组织/身份（如果适用）
	_, err = contract.SubmitTransaction("AddCertificate", string(certJsonBytes))                         // 赋值给现有的err
	fmt.Printf("DEBUG: AddCertificate 的 SubmitTransaction 已完成。原始错误: %v\n", err)                          // DEBUG LOG 5

	if err != nil {
		// 在删除文件并返回简化错误之前记录详细错误
		fmt.Printf("详细的链码调用错误 (AddCertificate raw error as seen by server): %v\n", err)
		// 如果链码提交失败，尝试删除保存的文件
		os.Remove(serverFilePath)
		return nil, fmt.Errorf("调用链码 AddCertificate 失败: %s", fabric.ExtractErrorMessage(err)) // 保留现有的面向用户的错误
	} else {
		fmt.Printf("DEBUG: 服务器认为 AddCertificate 的链码 SubmitTransaction 调用成功。文件应位于 %s。\n", serverFilePath) // DEBUG LOG 6
	}

	// 如果一切成功，err应该为nil
	return &certPayload, err // 返回最终的err状态（成功时应为nil）
}

// GetCertificatesByCar 通过过滤链码结果检索车辆的所有证书
func (s *CertificateService) GetCertificatesByCar(carId string) ([]*CertificatePayload, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG) // 或使用专用组织/身份（如果适用）

	resultBytes, err := contract.EvaluateTransaction("GetAllCertificates") // 此函数中err的首次声明
	if err != nil {
		errMsg := fabric.ExtractErrorMessage(err)
		if strings.Contains(errMsg, "不存在") || strings.Contains(strings.ToLower(errMsg), "not found") {
			return []*CertificatePayload{}, nil // 如果未找到证书，返回空列表
		}
		return nil, fmt.Errorf("调用链码 GetAllCertificates 失败: %s", errMsg)
	}

	if len(resultBytes) == 0 || string(resultBytes) == "null" || string(resultBytes) == "[]" {
		return []*CertificatePayload{}, nil
	}

	var allCertificates []*CertificatePayload
	err = json.Unmarshal(resultBytes, &allCertificates) // 赋值给现有的err
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

// GetCertificateFileLocation 检索给定证书ID的文件路径
func (s *CertificateService) GetCertificateFileLocation(certId string) (string, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)

	allCertsBytes, err := contract.EvaluateTransaction("GetAllCertificates") // 此函数中err的首次声明
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
	err = json.Unmarshal(allCertsBytes, &allCertificates) // 赋值给现有的err
	if err != nil {
		return "", fmt.Errorf("解析链码返回的证书列表失败: %v, Raw: %s", err, string(allCertsBytes))
	}

	for _, cert := range allCertificates {
		if cert != nil && cert.CertID == certId {
			// cert.FileLocation现在类似于"CAR_ID/filename.ext"
			// 重建完整的服务器路径
			fullServerPath := filepath.Join(certificateBaseDir, cert.FileLocation)
			absPath, err := filepath.Abs(fullServerPath) // 这里可以遮蔽err
			if err != nil {
				fmt.Printf("警告: 无法获取绝对路径 for %s (reconstructed from %s): %v\n", fullServerPath, cert.FileLocation, err)
				// 如果Abs失败但看起来已经是绝对路径，则回退到返回重建的路径
				// 或者只返回重建的路径
				if filepath.IsAbs(fullServerPath) {
					return fullServerPath, nil
				}
				return fullServerPath, nil // 返回从CWD重建的相对路径
			}
			return absPath, nil
		}
	}

	return "", fmt.Errorf("证书 %s 未找到 (not in list)", certId)
}

// VerifyCertificate 比较区块链中存储的哈希与服务器上实际文件的哈希
func (s *CertificateService) VerifyCertificate(certId string) (bool, string, string, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)
	resultBytes, err := contract.EvaluateTransaction("GetCertificate", certId) // 此函数中err的首次声明
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
	err = json.Unmarshal(resultBytes, &certPayload) // 赋值给现有的err
	if err != nil {
		return false, "", "", fmt.Errorf("解析链码返回的证书数据失败: %v, Raw: %s", err, string(resultBytes))
	}

	storedHash := certPayload.FileHash
	chaincodeFileLocation := certPayload.FileLocation // 这是"CAR_ID/filename.ext"

	if storedHash == "" || chaincodeFileLocation == "" {
		return false, "", "", fmt.Errorf("链上证书记录缺少哈希或文件路径信息")
	}

	// 从chaincodeFileLocation重建完整的服务器路径
	serverFilePath := filepath.Join(certificateBaseDir, chaincodeFileLocation)

	if _, err = os.Stat(serverFilePath); os.IsNotExist(err) { // 赋值给现有的err
		return false, storedHash, "", fmt.Errorf("服务器上找不到文件: %s (reconstructed from %s)", serverFilePath, chaincodeFileLocation)
	}

	file, err := os.Open(serverFilePath) // 赋值给现有的err，声明file
	if err != nil {
		return false, storedHash, "", fmt.Errorf("打开服务器文件 %s 失败: %v", serverFilePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file) // 赋值给现有的err
	if err != nil {
		return false, storedHash, "", fmt.Errorf("计算文件 %s 哈希失败: %v", serverFilePath, err)
	}
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	match := storedHash == currentHash

	return match, storedHash, currentHash, nil
}

// VerifyUploadedCertificate 比较上传文件的哈希与区块链上存储的给定车辆原始证书的哈希
func (s *CertificateService) VerifyUploadedCertificate(carId string, fileHeader *multipart.FileHeader) (bool, string, string, error) {
	originalCerts, err := s.GetCertificatesByCar(carId) // 此函数中err的首次声明
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

	uploadedFile, err := fileHeader.Open() // 赋值给现有的err，声明uploadedFile
	if err != nil {
		return false, storedHash, "", fmt.Errorf("打开待验证文件失败: %v", err)
	}
	defer uploadedFile.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, uploadedFile) // 赋值给现有的err
	if err != nil {
		return false, storedHash, "", fmt.Errorf("计算待验证文件哈希失败: %v", err)
	}
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	match := storedHash == currentHash

	return match, storedHash, currentHash, nil
}
