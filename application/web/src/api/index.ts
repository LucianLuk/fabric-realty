import request from '../utils/request';
import type { VehiclePageResult, TransactionPageResult, Vehicle, Transaction, BlockQueryResult } from '../types';

// 车辆管理机构接口
export const VehicleAgencyApi = {
  // 创建车辆信息
  createVehicle: (data: {
    id: string;
    model: string;
    brand: string;
    year: number;
    mileage: number;
    condition: string;
    owner: string;
  }) => request.post<never, void>('/vehicle-agency/vehicle/create', data),

  // 查询车辆信息
  getVehicleById: (id: string) => request.get<never, Vehicle>(`/vehicle-agency/vehicle/${id}`),

  // 分页查询车辆列表
  getVehicleList: (params: { pageSize: number; bookmark: string; status?: string }) =>
    request.get<never, VehiclePageResult>('/vehicle-agency/vehicle/list', { params }),

  // 分页查询区块列表
  getBlockList: (params: { pageSize?: number; pageNum?: number }) =>
    request.get<never, BlockQueryResult>('/vehicle-agency/block/list', { params }),
};

// 交易平台接口
export const tradingPlatformApi = {
  // 生成交易
  createTransaction: (data: {
    txId: string;
    vehicleId: string;
    seller: string;
    buyer: string;
    price: number;
  }) => request.post<never, void>('/trading-platform/transaction/create', data),

  // 查询车辆信息
  getVehicle: (id: string) => request.get<never, Vehicle>(`/trading-platform/vehicle/${id}`),

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
