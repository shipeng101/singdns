import axios from 'axios';

const TOKEN_KEY = 'singdns_token';

export const login = async (username, password) => {
  // 模拟登录成功，实际项目中应该调用后端 API
  if (username === 'admin' && password === 'admin') {
    const mockToken = 'mock_jwt_token';
    localStorage.setItem(TOKEN_KEY, mockToken);
    setAuthHeader(mockToken);
    return mockToken;
  }
  throw new Error('用户名或密码错误');
};

export const logout = () => {
  localStorage.removeItem(TOKEN_KEY);
  removeAuthHeader();
};

export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

export const setAuthHeader = (token) => {
  if (token) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }
};

export const removeAuthHeader = () => {
  delete axios.defaults.headers.common['Authorization'];
};

export const initAuth = () => {
  const token = getToken();
  if (token) {
    setAuthHeader(token);
  }
};

// 添加请求拦截器
axios.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 添加响应拦截器
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
); 