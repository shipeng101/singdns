import axios from 'axios';
import config from '../config';

// Set base URL for all requests
axios.defaults.baseURL = config.apiBaseUrl;

const TOKEN_KEY = 'singdns_token';

export const login = async (username, password) => {
  try {
    const response = await axios.post('/api/auth/login', { username, password });
    const { token } = response.data;
    localStorage.setItem(TOKEN_KEY, token);
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    return response.data;
  } catch (error) {
    throw new Error(error.response?.data?.error || '登录失败');
  }
};

export const logout = () => {
  localStorage.removeItem(TOKEN_KEY);
  delete axios.defaults.headers.common['Authorization'];
};

export const updatePassword = async (oldPassword, newPassword) => {
  try {
    const response = await axios.put('/api/user/password', { 
      old_password: oldPassword,
      new_password: newPassword 
    });
    return response.data;
  } catch (error) {
    throw new Error(error.response?.data?.error || '修改密码失败');
  }
};

export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

// Initialize axios auth header if token exists
const token = getToken();
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
} 