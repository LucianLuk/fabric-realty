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

// 车辆信息
export interface Vehicle {
  id: string;
  model: string;
  brand: string;
  year: number;
  mileage: number;
  condition: string;
  currentOwner: string;
  status: 'NORMAL' | 'IN_TRANSACTION';
  createTime: string;
  updateTime: string;
}

// 交易信息
export interface Transaction {
  id: string;
  vehicleId: string;
  seller: string;
  buyer: string;
  price: number;
  status: 'PENDING' | 'COMPLETED' | 'CANCELLED';
  createTime: string;
  updateTime: string;
}

// 车辆列表查询结果
export type VehiclePageResult = PageResult<Vehicle>;

// 交易列表查询结果
export type TransactionPageResult = PageResult<Transaction>; 