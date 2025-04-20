// API 响应基础结构
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 区块数据类型
export interface BlockData {
  block_num: number;
  block_hash: string;
  data_hash: string;
  prev_hash: string;
  tx_count: number;
  save_time: string;
}

// 区块查询结果类型
export interface BlockQueryResult {
  blocks: BlockData[];
  total: number;
  page_size: number;
  page_num: number;
  has_more: boolean;
}

// 分页查询结果
export interface PageResult<T> {
  bookmark: string;
  fetchedRecordsCount: number;
  records: T[];
  recordsCount: number;
}

// 汽车信息 (替代 RealEstate)
export interface Car {
  id: string;
  model: string; // 车型
  vin: string;   // 车辆识别代号
  currentOwner: string;
  status: 'AVAILABLE' | 'IN_TRANSACTION' | 'SOLD'; // 修改状态
  createTime: string;
  updateTime: string;
}

// 交易信息
export interface Transaction {
  id: string;
  carId: string; // 修改为 carId
  seller: string;
  buyer: string;
  price: number;
  status: 'PENDING' | 'COMPLETED' | 'CANCELLED';
  createTime: string;
  updateTime: string;
}

// 证书信息 (新增)
export interface Certificate {
  certId: string;
  carId: string;
  certType: string;
  fileHash: string;
  fileLocation: string; // Relative path from server data dir
  uploadTime: string; // ISO 8601 format string
}

// 汽车列表查询结果 (替代 RealEstatePageResult)
export type CarPageResult = PageResult<Car>;

// 交易列表查询结果
export type TransactionPageResult = PageResult<Transaction>;
