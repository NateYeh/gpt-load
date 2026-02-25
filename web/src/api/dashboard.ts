import type { ChartData, DashboardStatsResponse, Group } from "@/types/models";
import http from "@/utils/http";

/**
 * 获取仪表盘基础统计数据
 */
export const getDashboardStats = () => {
  return http.get<DashboardStatsResponse>("/dashboard/stats");
};

/**
 * 获取仪表盘图表数据
 * @param groupId 可选的分组ID
 * @param range 范围："24h" 或 "7d"
 */
export const getDashboardChart = (groupId?: number, range: "24h" | "7d" = "24h") => {
  return http.get<ChartData>("/dashboard/chart", {
    params: {
      groupId: groupId || undefined,
      range: range,
    },
  });
};

/**
 * 获取用于筛选的分组列表
 */
export const getGroupList = () => {
  return http.get<Group[]>("/groups/list");
};
