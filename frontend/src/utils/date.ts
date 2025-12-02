/**
 * 时间格式化工具函数
 * 所有时间都按照北京时间（UTC+8）显示
 */

/**
 * 格式化日期为北京时间
 * @param dateString ISO 8601 格式的日期字符串
 * @param options Intl.DateTimeFormatOptions 选项
 * @returns 格式化后的日期字符串
 */
export function formatDateToBeijing(
  dateString: string,
  options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }
): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('zh-CN', {
    ...options,
    timeZone: 'Asia/Shanghai',
  }).format(date);
}

/**
 * 格式化日期时间为北京时间（包含时间）
 * @param dateString ISO 8601 格式的日期字符串
 * @returns 格式化后的日期时间字符串，格式：YYYY-MM-DD HH:mm:ss
 */
export function formatDateTimeToBeijing(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
    timeZone: 'Asia/Shanghai',
  })
    .format(date)
    .replace(/\//g, '-');
}

/**
 * 格式化日期为北京时间（仅日期）
 * @param dateString ISO 8601 格式的日期字符串
 * @returns 格式化后的日期字符串，格式：YYYY-MM-DD
 */
export function formatDateOnlyToBeijing(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: 'Asia/Shanghai',
  })
    .format(date)
    .replace(/\//g, '-');
}

/**
 * 格式化相对时间（如：刚刚、5分钟前、2小时前等）
 * @param dateString ISO 8601 格式的日期字符串
 * @returns 相对时间字符串
 */
export function formatRelativeTimeToBeijing(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  
  // 转换为北京时间
  const beijingDate = new Date(
    date.toLocaleString('en-US', { timeZone: 'Asia/Shanghai' })
  );
  const beijingNow = new Date(
    now.toLocaleString('en-US', { timeZone: 'Asia/Shanghai' })
  );
  
  const diffMs = beijingNow.getTime() - beijingDate.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffSeconds < 60) {
    return '刚刚';
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`;
  } else if (diffHours < 24) {
    return `${diffHours}小时前`;
  } else if (diffDays < 7) {
    return `${diffDays}天前`;
  } else {
    return formatDateOnlyToBeijing(dateString);
  }
}

