import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Button,
  Typography,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  Stack,
  Alert,
  Snackbar,
  CircularProgress,
  LinearProgress,
  useTheme,
  FormControlLabel,
  Switch,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  FormHelperText,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Divider,
  Pagination,
  Checkbox,
  RadioGroup,
  FormControlLabel as MuiFormControlLabel,
  Radio,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Sync as SyncIcon,
  ViewList as ViewListIcon,
  Dashboard as DashboardIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { getNodes, getNodeGroups, createNodeGroup, updateNodeGroup, deleteNodeGroup, createNode, updateNode, deleteNode, testNode, testNodes, getSubscriptions } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Nodes = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [nodes, setNodes] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [openGroupDialog, setOpenGroupDialog] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [editGroup, setEditGroup] = useState({
    name: '',
    tag: '',
    includePatterns: '',
    excludePatterns: '',
    mode: 'select',
    active: true,
    matchMode: 'include'
  });
  const [openNodeDialog, setOpenNodeDialog] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [editNode, setEditNode] = useState({
    name: '',
    type: 'ss',
    address: '',
    port: '',
    method: 'aes-256-gcm',
    password: '',
    uuid: '',
    alterId: 0,
    security: 'auto',
    host: '',
    skipCertVerify: false,
    protocol: 'udp',
    up: '',
    down: '',
    obfs: '',
    alpn: [],
    upMbps: 100,
    downMbps: 100,
    version: 3,
    username: '',
    tls: false,
    network: 'tcp',
    udp: true,
    webSocket: false,
    path: '',
    wsHeaders: {},
    plugin: '',
    pluginOpts: {},
    cc: '',
    serviceName: ''
  });
  const [selectedNodes, setSelectedNodes] = useState([]);
  const [openNodeListDialog, setOpenNodeListDialog] = useState(false);
  const [selectedGroupNodes, setSelectedGroupNodes] = useState([]);
  const [selectedGroupForNodes, setSelectedGroupForNodes] = useState(null);

  const nodeTypes = [
    { value: 'ss', label: 'Shadowsocks' },
    { value: 'vmess', label: 'VMess' },
    { value: 'trojan', label: 'Trojan' },
    { value: 'naive', label: 'Naive' },
    { value: 'hysteria', label: 'Hysteria' },
    { value: 'hysteria2', label: 'Hysteria2' },
    { value: 'shadowtls', label: 'ShadowTLS' },
    { value: 'tun', label: 'Tun' },
    { value: 'redirect', label: 'Redirect' },
    { value: 'tproxy', label: 'TProxy' },
    { value: 'socks', label: 'Socks' },
    { value: 'http', label: 'HTTP' },
  ];

  const ssEncryption = [
    'aes-128-gcm',
    'aes-256-gcm',
    'chacha20-poly1305',
    'xchacha20-poly1305',
    '2022-blake3-aes-128-gcm',
    '2022-blake3-aes-256-gcm',
    '2022-blake3-chacha20-poly1305',
  ];

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [nodesResponse, groupsData, subscriptionsData] = await Promise.all([
        getNodes(page, pageSize),
        getNodeGroups(),
        getSubscriptions()
      ]);

      // Create a map of subscription id to subscription name
      const subscriptionMap = subscriptionsData.reduce((map, sub) => {
        map[sub.id] = sub.name;
        return map;
      }, {});

      // Add subscription name to nodes
      const updatedNodes = nodesResponse.nodes.map(node => {
        const checkedAt = localStorage.getItem(`node_${node.id}_checkedAt`);
        return {
          ...node,
          subscription_name: node.subscription_id ? subscriptionMap[node.subscription_id] : null,
          checkedAt: checkedAt || node.checkedAt
        };
      });

      // Sort nodes by latency (offline nodes at the end)
      const sortedNodes = updatedNodes.sort((a, b) => {
        if (a.status === 'offline' && b.status === 'offline') return 0;
        if (a.status === 'offline') return 1;
        if (b.status === 'offline') return -1;
        return a.latency - b.latency;
      });

      setNodes(sortedNodes);
      setTotal(nodesResponse.pagination.total);
      setNodeGroups(groupsData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleRefresh = async () => {
    try {
      setLoading(true);
      // 测试所有节点
      const response = await testNodes();
      // 更新节点状态
      const updatedNodes = nodes.map(node => {
        const result = response.results.find(r => r.id === node.id);
        if (result) {
          const checkedAt = result.checkedAt || new Date().toISOString();
          // 保存测速时间到 localStorage
          localStorage.setItem(`node_${node.id}_checkedAt`, checkedAt);
          return {
            ...node,
            status: result.status,
            latency: result.latency,
            checkedAt
          };
        }
        return node;
      });

      // 按延迟排序，离线节点放在最后
      const sortedNodes = updatedNodes.sort((a, b) => {
        if (a.status === 'offline' && b.status === 'offline') return 0;
        if (a.status === 'offline') return 1;
        if (b.status === 'offline') return -1;
        return a.latency - b.latency;
      });

      setNodes(sortedNodes);
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '刷新节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (event, newPage) => {
    setPage(newPage);
  };

  const handleOpenGroupDialog = (group = null) => {
    if (group) {
      setSelectedGroup(group);
      setEditGroup({
        ...group,
        includePatterns: Array.isArray(group.includePatterns) 
          ? group.includePatterns.join('\n')
          : '',
        excludePatterns: Array.isArray(group.excludePatterns)
          ? group.excludePatterns.join('\n')
          : '',
        mode: group.mode || 'select',
        tag: group.tag || '',
        active: group.active ?? true,
        matchMode: group.includePatterns?.length > 0 ? 'include' : 'exclude'
      });
    } else {
      setSelectedGroup(null);
      setEditGroup({
        name: '',
        tag: '',
        includePatterns: '',
        excludePatterns: '',
        mode: 'select',
        active: true,
        matchMode: 'include'
      });
    }
    setOpenGroupDialog(true);
  };

  const handleCloseGroupDialog = () => {
    setOpenGroupDialog(false);
    setSelectedGroup(null);
    setEditGroup({
      name: '',
      tag: '',
      includePatterns: '',
      excludePatterns: '',
      mode: 'select',
      active: true,
      matchMode: 'include'
    });
  };

  const handleSaveGroup = async () => {
    try {
      setLoading(true);
      setError('');

      // Validate required fields
      if (!editGroup.name?.trim()) {
        setError('分组名称不能为空');
        return;
      }

      if (!editGroup.tag?.trim()) {
        setError('分组标签不能为空');
        return;
      }

      // Check if name already exists in other groups
      const existingGroupWithName = nodeGroups.find(g => g.name === editGroup.name && g.id !== editGroup.id);
      if (existingGroupWithName) {
        setError('分组名称已存在，请使用其他名称');
        return;
      }

      // Check if tag already exists in other groups
      const existingGroupWithTag = nodeGroups.find(g => g.tag === editGroup.tag && g.id !== editGroup.id);
      if (existingGroupWithTag) {
        setError('分组标签已存在，请使用其他标签');
        return;
      }

      let includePatterns = [];
      let excludePatterns = [];

      // 根据匹配模式处理规则
      if (editGroup.name === "全部") {
        // "全部"分组使用特殊规则
        includePatterns = [".*"];  // 匹配所有节点
        excludePatterns = [];      // 不排除任何节点
      } else if (editGroup.matchMode === 'include') {
        // 处理白名单规则
        if (editGroup.includePatterns) {
          // 先按分割，并过滤空行
          const lines = editGroup.includePatterns
            .split('\n')
            .map(line => line.trim())
            .filter(line => line !== '' && line !== '\n');
          
          // 如果有输入规则，使用输入的规则
          if (lines.length > 0) {
            includePatterns = lines;
          }
        }
        
        // 如果没有规则，使用分组名称作为关键字
        if (includePatterns.length === 0) {
          includePatterns = [editGroup.name];
        }
      } else {
        // 黑名单模式
        includePatterns = ['.*'];  // 默认包含所有节点
        if (editGroup.excludePatterns) {
          excludePatterns = editGroup.excludePatterns
            .split('\n')
            .map(line => line.trim())
            .filter(line => line !== '' && line !== '\n');
        }
      }

      const groupData = {
        ...editGroup,
        include_patterns: includePatterns,
        exclude_patterns: excludePatterns,
        active: editGroup.active ?? true,
        mode: editGroup.mode || 'select'
      };

      if (selectedGroup) {
        await updateNodeGroup(selectedGroup.id, groupData);
      } else {
        await createNodeGroup(groupData);
      }
      
      await new Promise(resolve => setTimeout(resolve, 1000));
      await fetchData();
      setSuccess(true);
      handleCloseGroupDialog();
    } catch (err) {
      console.error('Error saving group:', err);
      setError(err.response?.data?.error || '保存分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteGroup = async (id) => {
    try {
      setLoading(true);
      await deleteNodeGroup(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenNodeDialog = (node = null) => {
    if (node) {
      setSelectedNode(node);
      setEditNode({
        ...node,
        webSocket: node.network === 'ws',
      });
    } else {
      setSelectedNode(null);
      setEditNode({
        name: '',
        type: 'ss',
        address: '',
        port: '',
        method: 'aes-256-gcm',
        password: '',
        uuid: '',
        alterId: 0,
        security: 'auto',
        host: '',
        skipCertVerify: false,
        protocol: 'udp',
        up: '',
        down: '',
        obfs: '',
        alpn: [],
        upMbps: 100,
        downMbps: 100,
        version: 3,
        username: '',
        tls: false,
        network: 'tcp',
        udp: true,
        webSocket: false,
        path: '',
        wsHeaders: {},
        plugin: '',
        pluginOpts: {},
        cc: '',
        serviceName: ''
      });
    }
    setOpenNodeDialog(true);
  };

  const handleCloseNodeDialog = () => {
    setOpenNodeDialog(false);
    setSelectedNode(null);
    setEditNode({
      name: '',
      type: 'ss',
      address: '',
      port: '',
      method: 'aes-256-gcm',
      password: '',
      uuid: '',
      alterId: 0,
      security: 'auto',
      host: '',
      skipCertVerify: false,
      protocol: 'udp',
      up: '',
      down: '',
      obfs: '',
      alpn: [],
      upMbps: 100,
      downMbps: 100,
      version: 3,
      username: '',
      tls: false,
      network: 'tcp',
      udp: true,
      webSocket: false,
      path: '',
      wsHeaders: {},
      plugin: '',
      pluginOpts: {},
      cc: '',
      serviceName: ''
    });
  };

  const handleSaveNode = async () => {
    try {
      setLoading(true);
      const nodeData = {
        ...editNode,
        id: selectedNode ? selectedNode.id : undefined,
        group: selectedNode ? selectedNode.group : '',
        status: selectedNode ? selectedNode.status : 'offline',
        latency: selectedNode ? selectedNode.latency : 0,
        checkedAt: selectedNode ? selectedNode.checkedAt : new Date().toISOString(),
        createdAt: selectedNode ? selectedNode.createdAt : Math.floor(Date.now() / 1000),
        updatedAt: Math.floor(Date.now() / 1000)
      };

      if (selectedNode) {
        await updateNode(selectedNode.id, nodeData);
      } else {
        await createNode(nodeData);
      }
      await fetchData();
      setSuccess(true);
      handleCloseNodeDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNode = async (id) => {
    if (!window.confirm('确定要删除这个节点吗？')) {
      return;
    }
    
    try {
      setLoading(true);
      await deleteNode(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectAll = (event) => {
    if (event.target.checked) {
      setSelectedNodes(nodes.map(node => node.id));
    } else {
      setSelectedNodes([]);
    }
  };

  const handleSelectNode = (nodeId) => {
    setSelectedNodes(prev => {
      if (prev.includes(nodeId)) {
        return prev.filter(id => id !== nodeId);
      } else {
        return [...prev, nodeId];
      }
    });
  };

  const handleBatchDelete = async () => {
    if (selectedNodes.length === 0) {
      setError('请先选择要删除的节点');
      return;
    }

    if (!window.confirm(`确定要删除选中的 ${selectedNodes.length} 个节点吗？`)) {
      return;
    }

    try {
      setLoading(true);
      for (const nodeId of selectedNodes) {
        await deleteNode(nodeId);
      }
      await fetchData();
      setSelectedNodes([]);
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '批量删除节点失败');
    } finally {
      setLoading(false);
    }
  };

  const renderNodeFields = () => {
    switch (editNode.type) {
      case 'ss':
        return (
          <>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.method}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, method: e.target.value })}
                  sx={styles.textField}
                >
                  {ssEncryption.map((method) => (
                    <MenuItem key={method} value={method}>{method}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="插件"
                value={editNode.plugin}
                onChange={(e) => setEditNode({ ...editNode, plugin: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="插件选项"
                value={editNode.pluginOpts}
                onChange={(e) => setEditNode({ ...editNode, pluginOpts: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
          </>
        );
      case 'vmess':
        return (
          <>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="UUID"
                value={editNode.uuid}
                onChange={(e) => setEditNode({ ...editNode, uuid: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="AlterID"
                type="number"
                value={editNode.alterId}
                onChange={(e) => setEditNode({ ...editNode, alterId: parseInt(e.target.value) || 0 })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.security}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, security: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value="auto">auto</MenuItem>
                  <MenuItem value="aes-128-gcm">aes-128-gcm</MenuItem>
                  <MenuItem value="chacha20-poly1305">chacha20-poly1305</MenuItem>
                  <MenuItem value="none">none</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>传输协议</InputLabel>
                <Select
                  value={editNode.network}
                  label="传输协议"
                  onChange={(e) => setEditNode({ ...editNode, network: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value="tcp">TCP</MenuItem>
                  <MenuItem value="ws">WebSocket</MenuItem>
                  <MenuItem value="http">HTTP</MenuItem>
                  <MenuItem value="h2">HTTP/2</MenuItem>
                  <MenuItem value="grpc">gRPC</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            {editNode.network === 'ws' && (
              <>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="WebSocket 路径"
                    value={editNode.path}
                    onChange={(e) => setEditNode({ ...editNode, path: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="WebSocket 主机"
                    value={editNode.host}
                    onChange={(e) => setEditNode({ ...editNode, host: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
              </>
            )}
          </>
        );
      case 'trojan':
        return (
          <>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="SNI"
                value={editNode.host}
                onChange={(e) => setEditNode({ ...editNode, host: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
          </>
        );
      default:
        return null;
    }
  };

  // 添加 URL 码函数
  const decodeNodeName = (name) => {
    try {
      return decodeURIComponent(name);
    } catch (e) {
      return name;
    }
  };

  // 格式化延时显示
  const formatLatency = (latency) => {
    if (latency === undefined || latency === null) return '-';
    if (latency === 0) return '-';
    if (latency < 0) return '超时';
    return `${latency}ms`;
  };

  // 获取节点状态的颜色
  const getStatusColor = (status, latency) => {
    if (status === 'online') {
      if (latency < 100) {
        return 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)';  // 绿色
      } else if (latency < 300) {
        return 'linear-gradient(45deg, #ff9800 30%, #ffb74d 90%)';  // 橙色
      } else {
        return 'linear-gradient(45deg, #f44336 30%, #e57373 90%)';  // 红色
      }
    }
    return 'linear-gradient(45deg, #9e9e9e 30%, #bdbdbd 90%)';  // 灰色
  };

  // 添加单个节点更新函数
  const handleRefreshNode = async (nodeId) => {
    try {
      setLoading(true);
      // 测试单个节点
      const response = await testNodes([nodeId]);
      // 更新节点状态
      setNodes(prevNodes => {
        return prevNodes.map(node => {
          if (node.id === nodeId) {
            const result = response.results.find(r => r.id === nodeId);
            if (result) {
              const checkedAt = result.checkedAt || new Date().toISOString();
              // 保存测速时间到 localStorage
              localStorage.setItem(`node_${node.id}_checkedAt`, checkedAt);
              return {
                ...node,
                status: result.status,
                latency: result.latency,
                checkedAt
              };
            }
          }
          return node;
        }).sort((a, b) => {
          if (a.status === 'offline' && b.status === 'offline') return 0;
          if (a.status === 'offline') return 1;
          if (b.status === 'offline') return -1;
          return a.latency - b.latency;
        });
      });
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '刷新节点失败');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    { field: 'name', headerName: '名称', flex: 1 },
    { field: 'type', headerName: '类型', width: 100 },
    { field: 'address', headerName: '地址', flex: 1 },
    { field: 'port', headerName: '端口', width: 100 },
    {
      field: 'status',
      headerName: '状态',
      width: 120,
      renderCell: (params) => (
        <Chip
          label={params.value === 'online' ? '在线' : '离线'}
          color={params.value === 'online' ? 'success' : 'error'}
          size="small"
          sx={{
            background: getStatusColor(params.value, params.row.latency),
            color: 'white'
          }}
        />
      ),
    },
    {
      field: 'latency',
      headerName: '延迟',
      width: 100,
      renderCell: (params) => (
        <span>{formatLatency(params.value)}</span>
      ),
    },
    {
      field: 'checkedAt',
      headerName: '最后测速',
      width: 180,
      renderCell: (params) => (
        params.value ? new Date(params.value).toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false
        }) : '-'
      ),
    },
    {
      field: 'actions',
      headerName: '操作',
      width: 120,
      renderCell: (params) => (
        <Stack direction="row" spacing={1}>
          <IconButton 
            size="small" 
            onClick={() => handleRefreshNode(params.row.id)}
            sx={{
              color: 'text.secondary',
              '&:hover': {
                color: 'primary.main',
                backgroundColor: 'action.hover',
              },
            }}
          >
            <SyncIcon fontSize="small" />
          </IconButton>
          <IconButton 
            size="small" 
            onClick={() => handleOpenNodeDialog(params.row)}
            sx={{
              color: 'text.secondary',
              '&:hover': {
                color: 'primary.main',
                backgroundColor: 'action.hover',
              },
            }}
          >
            <EditIcon fontSize="small" />
          </IconButton>
          <IconButton 
            size="small" 
            onClick={() => handleDeleteNode(params.row.id)}
            sx={{
              color: 'text.secondary',
              '&:hover': {
                color: 'error.main',
                backgroundColor: 'action.hover',
              },
            }}
          >
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Stack>
      ),
    },
  ];

  // 修复中文编码题
  const renderGroupModeHelperText = (mode) => {
    return mode === 'select' ? '手动选择节点' : '自动选择可用和延迟较低的节点';
  };

  // 渲染分组卡片
  const renderGroupCard = (group) => (
    <Card
      key={group.id}
      component={motion.div}
      whileHover={{ scale: 1.02 }}
      sx={styles.card}
    >
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
          <Typography variant="h6" component="div">
            {group.name}
          </Typography>
          <Box>
            <IconButton
              size="small"
              onClick={() => handleOpenNodeList(group)}
              sx={{ mr: 1 }}
            >
              <ViewListIcon fontSize="small" />
            </IconButton>
            <IconButton
              size="small"
              onClick={() => handleOpenGroupDialog(group)}
              sx={{ mr: 1 }}
            >
              <EditIcon fontSize="small" />
            </IconButton>
            <IconButton
              size="small"
              onClick={() => handleDeleteGroup(group.id)}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Box>
        </Box>
        <Typography variant="body2" color="text.secondary">
          {group.node_count || 0} 个节点
        </Typography>
        {group.tag && (
          <Chip
            label={group.tag}
            size="small"
            sx={{ mt: 1 }}
          />
        )}
      </CardContent>
    </Card>
  );

  // 添加查看节点列表的处理函数
  const handleOpenNodeList = async (group) => {
    try {
      setLoading(true);
      // 获取最新的节点组数据和订阅数据
      const [groupData, subscriptionsData] = await Promise.all([
        getNodeGroups(),
        getSubscriptions()
      ]);

      // 创建订阅 ID 到名称的映射
      const subscriptionMap = subscriptionsData.reduce((map, sub) => {
        map[sub.id] = sub.name;
        return map;
      }, {});

      const targetGroup = groupData.find(g => g.id === group.id);
      
      if (targetGroup) {
        // 过滤掉不存在的节点，并添加订阅名称
        const validNodes = (targetGroup.nodes || [])
          .filter(node => node && node.id)
          .map(node => ({
            ...node,
            subscription_name: node.subscription_id ? subscriptionMap[node.subscription_id] : null
          }));

        setSelectedGroupNodes(validNodes);
        setSelectedGroupForNodes(targetGroup);
        setOpenNodeListDialog(true);
      }
    } catch (err) {
      setError(err.response?.data?.error || '获取节点列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 添加从分组中移除节点的处理函数
  const handleRemoveNodeFromGroup = async (nodeId) => {
    try {
      setLoading(true);
      // 获取最新的节点组数据
      const groupData = await getNodeGroups();
      const currentGroup = groupData.find(g => g.id === selectedGroupForNodes.id);
      
      if (!currentGroup) {
        throw new Error('分组不存在');
      }

      // 过滤掉要移除的节点
      const updatedNodes = (currentGroup.nodes || []).filter(node => node.id !== nodeId);
      
      // 更新节点组数据
      await updateNodeGroup(selectedGroupForNodes.id, {
        ...currentGroup,
        nodes: updatedNodes
      });

      // 更新节点状态
      setSelectedGroupNodes(updatedNodes);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '移除节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleManualRefresh = async () => {
    await fetchData();
  };

  return (
    <Box>
      <Snackbar 
        open={success} 
        autoHideDuration={3000} 
        onClose={() => setSuccess(false)}
      >
        <Alert severity="success">操作成功</Alert>
      </Snackbar>

      <Snackbar 
        open={!!error} 
        autoHideDuration={3000} 
        onClose={() => setError(null)}
      >
        <Alert severity="error">{error}</Alert>
      </Snackbar>

      <Card sx={styles.headerCard}>
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={2}>
              <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
                节点管理
              </Typography>
              {loading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
              <IconButton
                size="small"
                onClick={handleManualRefresh}
                disabled={loading}
                sx={styles.iconButton}
                title="刷新"
              >
                <SyncIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
                title="打开仪表盘"
              >
                <DashboardIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
                title={mode === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      {/* Nodes List */}
      <Box sx={{ mb: 3 }}>
        <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">节点列表</Typography>
          <Stack direction="row" spacing={1}>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => handleOpenNodeDialog()}
              sx={styles.actionButton}
            >
              添加节点
            </Button>
            <Button
              variant="contained"
              size="small"
              startIcon={<SyncIcon />}
              onClick={handleRefresh}
              disabled={loading}
              sx={styles.actionButton}
            >
              刷新节点
            </Button>
            {selectedNodes.length > 0 && (
              <Button
                variant="contained"
                size="small"
                color="error"
                startIcon={<DeleteIcon />}
                onClick={handleBatchDelete}
                disabled={loading}
                sx={styles.actionButton}
              >
                删除选中 ({selectedNodes.length})
              </Button>
            )}
          </Stack>
        </Box>

        <TableContainer component={Paper} sx={styles.tableContainer}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell padding="checkbox">
                  <Checkbox
                    indeterminate={selectedNodes.length > 0 && selectedNodes.length < nodes.length}
                    checked={nodes.length > 0 && selectedNodes.length === nodes.length}
                    onChange={handleSelectAll}
                    size="small"
                  />
                </TableCell>
                <TableCell>名称</TableCell>
                <TableCell>类型</TableCell>
                <TableCell>地址</TableCell>
                <TableCell>端口</TableCell>
                <TableCell>延迟</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>订阅</TableCell>
                <TableCell>最后测速</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {nodes.map((node) => (
                <TableRow key={node.id}>
                  <TableCell padding="checkbox">
                    <Checkbox
                      checked={selectedNodes.includes(node.id)}
                      onChange={() => handleSelectNode(node.id)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{decodeNodeName(node.name)}</TableCell>
                  <TableCell>{node.type}</TableCell>
                  <TableCell>{node.address}</TableCell>
                  <TableCell>{node.port}</TableCell>
                  <TableCell>
                    <Chip
                      label={formatLatency(node.latency)}
                      size="small"
                      sx={{
                        background: getStatusColor(node.status, node.latency),
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={node.status === 'online' ? '在线' : '离线'} 
                      size="small"
                      sx={{
                        background: node.status === 'online' 
                          ? 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)'
                          : 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </TableCell>
                  <TableCell>
                    {node.subscription_id ? (
                      <Chip
                        label={node.subscription_name}
                        size="small"
                        sx={{
                          backgroundColor: 'action.selected',
                          color: 'text.primary',
                        }}
                      />
                    ) : (
                      <Chip
                        label="手动添加"
                        size="small"
                        sx={{
                          backgroundColor: 'action.hover',
                          color: 'text.secondary',
                        }}
                      />
                    )}
                  </TableCell>
                  <TableCell>
                    {node.checkedAt ? new Date(node.checkedAt).toLocaleString('zh-CN', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit',
                      second: '2-digit',
                      hour12: false
                    }) : '-'}
                  </TableCell>
                  <TableCell align="right">
                    <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                      <IconButton 
                        size="small" 
                        onClick={() => handleRefreshNode(node.id)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'primary.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <SyncIcon fontSize="small" />
                      </IconButton>
                      <IconButton 
                        size="small" 
                        onClick={() => handleOpenNodeDialog(node)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'primary.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton 
                        size="small" 
                        onClick={() => handleDeleteNode(node.id)}
                        sx={{
                          color: 'text.secondary',
                          '&:hover': {
                            color: 'error.main',
                            backgroundColor: 'action.hover',
                          },
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
          <Pagination
            count={Math.ceil(total / pageSize)}
            page={page}
            onChange={handlePageChange}
            color="primary"
            size="small"
          />
        </Box>
      </Box>

      {/* Node Groups */}
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">节点分组</Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenGroupDialog()}
          sx={styles.actionButton}
        >
          添加分组
        </Button>
      </Box>

      <Grid container spacing={1.5}>
        {nodeGroups.map((group) => (
          <Grid item xs={12} sm={6} md={3} key={group.id}>
            {renderGroupCard(group)}
          </Grid>
        ))}
      </Grid>

      {/* Group Dialog */}
      <Dialog 
        open={openGroupDialog} 
        onClose={handleCloseGroupDialog} 
        maxWidth="sm" 
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedGroup ? '编辑分组' : '添加分组'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  required
                  label="分组名称"
                  value={editGroup.name}
                  onChange={(e) => setEditGroup({ ...editGroup, name: e.target.value })}
                  size="small"
                  placeholder="例如：香港节点, 日本节点"
                  helperText="此名称将显示在配置文件中，必须唯一"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  required
                  label="分组标签"
                  value={editGroup.tag}
                  onChange={(e) => setEditGroup({ ...editGroup, tag: e.target.value })}
                  size="small"
                  placeholder="例如：HK, JP, SG, US"
                  helperText="此标签将用作配置文件中的分组标识，必须唯一"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>节点选择模式</InputLabel>
                  <Select
                    value={editGroup.mode}
                    label="节点选择模式"
                    onChange={(e) => setEditGroup({ ...editGroup, mode: e.target.value })}
                    sx={styles.textField}
                  >
                    <MenuItem value="select">手动选择</MenuItem>
                    <MenuItem value="urltest">自动测速</MenuItem>
                  </Select>
                  <FormHelperText>
                    {renderGroupModeHelperText(editGroup.mode)}
                  </FormHelperText>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editGroup.active}
                      onChange={(e) => setEditGroup({ ...editGroup, active: e.target.checked })}
                    />
                  }
                  label="活跃"
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="subtitle2" gutterBottom>
                  节点匹配规则
                </Typography>
                <FormControl component="fieldset">
                  <RadioGroup
                    value={editGroup.matchMode}
                    onChange={(e) => setEditGroup({ ...editGroup, matchMode: e.target.value })}
                  >
                    <FormControlLabel 
                      value="include" 
                      control={<Radio />} 
                      label="白名单模式（包含指定节点）" 
                    />
                    <FormControlLabel 
                      value="exclude" 
                      control={<Radio />} 
                      label="黑名单模式（排除指定节点）" 
                    />
                  </RadioGroup>
                </FormControl>
                <Box sx={{ mt: 2 }}>
                  {editGroup.matchMode === 'include' ? (
                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        包含节点规则（白名单）
                      </Typography>
                      <TextField
                        fullWidth
                        multiline
                        rows={3}
                        placeholder="每行一个规则，支持两种格式：
1. 正则表达式（以 ^ 开头）：
^(香港|HK).*$
^(日本|JP).*$

2. 关键字匹配：
香港|HK
日本|JP
新加坡|SG|🇸🇬"
                        value={editGroup.includePatterns}
                        onChange={(e) => {
                          setEditGroup(prev => ({
                            ...prev,
                            includePatterns: e.target.value
                          }));
                        }}
                        size="small"
                        sx={styles.textField}
                      />
                    </Box>
                  ) : (
                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        排除节点规则（黑名单）
                      </Typography>
                      <TextField
                        fullWidth
                        multiline
                        rows={3}
                        placeholder="每行一个规则，格式同上：
1. 正则表达式：^.*过期.*$
2. 关键字：过期|测试|expire"
                        value={editGroup.excludePatterns}
                        onChange={(e) => {
                          setEditGroup(prev => ({
                            ...prev,
                            excludePatterns: e.target.value
                          }));
                        }}
                        size="small"
                        sx={styles.textField}
                      />
                    </Box>
                  )}
                </Box>
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseGroupDialog}
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            取消
          </Button>
          <Button 
            variant="contained" 
            onClick={handleSaveGroup}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* Node Dialog */}
      <Dialog
        open={openNodeDialog}
        onClose={handleCloseNodeDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedNode ? '编辑节点' : '添加节点'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="名称"
                  value={editNode.name}
                  onChange={(e) => setEditNode({ ...editNode, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>类型</InputLabel>
                  <Select
                    value={editNode.type}
                    label="类型"
                    onChange={(e) => setEditNode({ ...editNode, type: e.target.value })}
                    sx={styles.textField}
                  >
                    {nodeTypes.map((type) => (
                      <MenuItem key={type.value} value={type.value}>{type.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="地址"
                  value={editNode.address}
                  onChange={(e) => setEditNode({ ...editNode, address: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="端口"
                  type="number"
                  value={editNode.port}
                  onChange={(e) => setEditNode({ ...editNode, port: parseInt(e.target.value) || '' })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              {renderNodeFields()}
              <Grid item xs={12}>
                <Divider sx={{ my: 1 }}>高级设置</Divider>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.tls}
                      onChange={(e) => setEditNode({ ...editNode, tls: e.target.checked })}
                    />
                  }
                  label="启用 TLS"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.skipCertVerify}
                      onChange={(e) => setEditNode({ ...editNode, skipCertVerify: e.target.checked })}
                    />
                  }
                  label="跳过证书验证"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.udp}
                      onChange={(e) => setEditNode({ ...editNode, udp: e.target.checked })}
                    />
                  }
                  label="启用 UDP"
                />
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseNodeDialog}
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            取消
          </Button>
          <Button 
            variant="contained" 
            onClick={handleSaveNode}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* 添加节点列表对话框 */}
      <Dialog
        open={openNodeListDialog}
        onClose={() => setOpenNodeListDialog(false)}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedGroupForNodes?.name} - 节点列表
        </DialogTitle>
        <DialogContent>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>名称</TableCell>
                  <TableCell>类型</TableCell>
                  <TableCell>地址</TableCell>
                  <TableCell>延迟</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell>订阅</TableCell>
                  <TableCell align="right">操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {selectedGroupNodes.map((node) => (
                  <TableRow key={node.id}>
                    <TableCell>{decodeNodeName(node.name)}</TableCell>
                    <TableCell>{node.type}</TableCell>
                    <TableCell>{node.address}</TableCell>
                    <TableCell>
                      <Chip
                        label={formatLatency(node.latency)}
                        size="small"
                        sx={{
                          background: getStatusColor(node.status, node.latency),
                          color: '#fff',
                          fontWeight: 500,
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={node.status === 'online' ? '在线' : '离线'}
                        size="small"
                        sx={{
                          background: getStatusColor(node.status, node.latency),
                          color: '#fff',
                          fontWeight: 500,
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      {node.subscription_id ? (
                        <Chip
                          label={node.subscription_name}
                          size="small"
                          sx={{
                            backgroundColor: 'action.selected',
                            color: 'text.primary',
                          }}
                        />
                      ) : (
                        <Chip
                          label="手动添加"
                          size="small"
                          sx={{
                            backgroundColor: 'action.hover',
                            color: 'text.secondary',
                          }}
                        />
                      )}
                    </TableCell>
                    <TableCell align="right">
                      <Stack direction="row" spacing={1} justifyContent="flex-end">
                        <IconButton
                          size="small"
                          onClick={() => handleRefreshNode(node.id)}
                          sx={{
                            color: 'text.secondary',
                            '&:hover': {
                              color: 'primary.main',
                              backgroundColor: 'action.hover',
                            },
                          }}
                        >
                          <SyncIcon fontSize="small" />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => handleRemoveNodeFromGroup(node.id)}
                          sx={{
                            color: 'text.secondary',
                            '&:hover': {
                              color: 'error.main',
                              backgroundColor: 'action.hover',
                            },
                          }}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Stack>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => setOpenNodeListDialog(false)}
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            关闭
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Nodes; 