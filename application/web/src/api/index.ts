import request from '../utils/request';
// 修改导入的类型
import type { CarPageResult, TransactionPageResult, Car, Transaction, BlockQueryResult, Certificate } from '../types'; // Import Certificate

// 汽车经销商接口 (替代 realtyAgencyApi)
export const carDealerApi = {
  // 创建汽车信息
  createCar: (data: {
    id: string;
    model: string; // 修改字段
    vin: string;   // 修改字段
    owner: string;
  }) => request.post<never, void>('/car-dealer/car/create', data), // 修改路径

  // 查询汽车信息
  getCar: (id: string) => request.get<never, Car>(`/car-dealer/car/${id}`), // 修改路径和返回类型

  // 分页查询汽车列表
  getCarList: (params: { pageSize: number; bookmark: string; status?: string }) =>
    request.get<never, CarPageResult>('/car-dealer/car/list', { params }), // 修改路径和返回类型

  // 分页查询区块列表 (路径保持一致，但属于 car-dealer)
  getBlockList: (params: { pageSize?: number; pageNum?: number }) =>
    request.get<never, BlockQueryResult>('/car-dealer/block/list', { params }),

  // 上传证书 (新增)
  uploadCertificate: (carId: string, certType: string, file: File) => {
    const formData = new FormData();
    formData.append('certType', certType);
    formData.append('certificateFile', file); // Match the backend expected field name

    // 修改路径以匹配后端
    return request.post<never, Certificate>(`/car-dealer/certificates/${carId}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data', // Important for file uploads
      },
    });
  },

  // 获取车辆证书列表 (新增)
  listCertificates: (carId: string) =>
    // 修改路径以匹配后端
    request.get<never, Certificate[]>(`/car-dealer/certificates/${carId}`),

  // 验证证书完整性 (新增) - 验证服务器存储的文件
  verifyCertificate: (certId: string) =>
    request.get<never, { match: boolean; storedHash: string; currentHash: string }>(`/car-dealer/certificates/verify/${certId}`), // 修改路径

  // 上传文件进行验证 (新增) - 对比上传文件和原始文件
  verifyUploadedCertificate: (carId: string, file: File) => {
    const formData = new FormData();
    formData.append('verificationFile', file); // Match backend expected field name
    return request.post<never, { match: boolean; storedHash: string; currentHash: string }>(
      `/car-dealer/certificates/verify-upload/${carId}`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
  },
};

// 交易平台接口
export const tradingPlatformApi = {
  // 生成交易
  createTransaction: (data: {
    txId: string;
    carId: string; // 修改字段
    seller: string;
    buyer: string;
    price: number;
  }) => request.post<never, void>('/trading-platform/transaction/create', data),

  // 查询汽车信息 (替代 getRealEstate)
  getCar: (id: string) => request.get<never, Car>(`/trading-platform/car/${id}`), // 修改路径和返回类型

  // 查询交易信息
  getTransaction: (txId: string) => request.get<never, Transaction>(`/trading-platform/transaction/${txId}`),

  // 分页查询交易列表
  getTransactionList: (params: { pageSize: number; bookmark: string; status?: string }) =>
    request.get<never, TransactionPageResult>('/trading-platform/transaction/list', { params }),

  // 分页查询区块列表
  getBlockList: (params: { pageSize?: number; pageNum?: number }) =>
    request.get<never, BlockQueryResult>('/trading-platform/block/list', { params }),
};

// 银行接口
export const bankApi = {
  // 完成交易
  completeTransaction: (txId: string) =>
    request.post<never, void>(`/bank/transaction/complete/${txId}`),

  // 查询交易信息
  getTransaction: (txId: string) => request.get<never, Transaction>(`/bank/transaction/${txId}`),

  // 分页查询交易列表
  getTransactionList: (params: { pageSize: number; bookmark: string; status?: string }) =>
    request.get<never, TransactionPageResult>('/bank/transaction/list', { params }),

  // 分页查询区块列表
  getBlockList: (params: { pageSize?: number; pageNum?: number }) =>
    request.get<never, BlockQueryResult>('/bank/block/list', { params }),
};
