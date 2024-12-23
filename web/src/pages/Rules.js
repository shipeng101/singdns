import React, { useState, useEffect } from 'react';
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
  Chip,
  FormLabel,
  RadioGroup,
  FormControlLabel as MuiFormControlLabel,
  Radio
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Refresh as RefreshIcon
} from '@mui/icons-material';
import { getRules, createRule, updateRule, deleteRule, getNodes, getNodeGroups, getRuleSets, createRuleSet, updateRuleSet, deleteRuleSet, updateRuleSetRules } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';
import { motion } from 'framer-motion';

const Rules = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [rules, setRules] = useState([]);
  const [nodes, setNodes] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [editRule, setEditRule] = useState({
    name: '',
    type: 'custom',
    url: '',
    rules: [],
    outbound: 'direct',
    nodeGroup: '',
    specificNode: '',
    enabled: true,
    format: 'domain',
    description: '',
  });
  const [ruleSets, setRuleSets] = useState([]);

  const outboundTypes = [
    { value: 'direct', label: '直连', description: '不走代理直接连接' },
    { value: 'reject', label: '拒绝', description: '通过DNS拒绝连接' },
    { value: 'node_group', label: '节点组', description: '使用指定节点组' },
    { value: 'specific', label: '指定节点', description: '使用特定节点' },
  ];

  const ruleTypes = [
    { value: 'custom', label: '自定义规则', description: '手动添加规则' },
    { value: 'ruleset', label: '远程规则集', description: '导入SRS或JSON格式的规则集' },
  ];

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [rulesData, nodesData, groupsData, ruleSetsData] = await Promise.all([
        getRules(),
        getNodes(),
        getNodeGroups(),
        getRuleSets()
      ]);
      setRules(rulesData);
      setNodes(nodesData);
      setNodeGroups(groupsData);
      setRuleSets(ruleSetsData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (rule = null) => {
    if (rule) {
      setSelectedRule(rule);
      setEditRule({
        ...rule,
        type: rule.type === 'list' || rule.type === 'json' ? 'ruleset' : 'custom',
        format: rule.type === 'list' || rule.type === 'json' ? rule.type : rule.type || 'domain',
        ruleFormat: rule.format || 'domain',
        rules: rule.domains || rule.ips || [],
        enabled: rule.enabled ?? true,
      });
    } else {
      setSelectedRule(null);
      setEditRule({
        name: '',
        type: 'custom',
        format: 'domain',
        ruleFormat: 'domain',
        rules: [],
        url: '',
        outbound: 'direct',
        nodeGroup: '',
        specificNode: '',
        enabled: true,
        description: '',
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedRule(null);
    setEditRule({
      name: '',
      type: 'custom',
      format: 'domain',
      ruleFormat: 'domain',
      rules: [],
      url: '',
      outbound: 'direct',
      nodeGroup: '',
      specificNode: '',
      enabled: true,
      description: '',
    });
  };

  const handleSaveRule = async () => {
    try {
      setLoading(true);
      if (editRule.type === 'ruleset') {
        // 保存为规则集
        const ruleSetData = {
          name: editRule.name,
          url: editRule.url,
          type: editRule.format,  // 使用 format (srs/json) 作为类型
          format: editRule.ruleFormat || 'domain',  // 规则格式（域名/IP）
          outbound: editRule.outbound,
          description: editRule.description,
          enabled: editRule.enabled,
        };

        if (selectedRule) {
          if (selectedRule.type === 'srs' || selectedRule.type === 'json') {
            // 如果原来就是规则集，直接更新
            await updateRuleSet(selectedRule.id, ruleSetData);
          } else {
            // 如果是从普通规则转换为规则集，先删除原规则，再创建新规则集
            await deleteRule(selectedRule.id);
            await createRuleSet(ruleSetData);
          }
        } else {
          await createRuleSet(ruleSetData);
        }
      } else {
        // 保存为普通规则
        const ruleData = {
          name: editRule.name,
          type: editRule.format,
          domains: editRule.format === 'domain' ? editRule.rules : [],
          ips: editRule.format === 'geoip' ? editRule.rules : [],
          outbound: editRule.outbound,
          nodeGroup: editRule.nodeGroup,
          specificNode: editRule.specificNode,
          enabled: editRule.enabled,
          description: editRule.description,
        };

        if (selectedRule) {
          if (selectedRule.type === 'srs' || selectedRule.type === 'json') {
            // 如果是从规则集转换为普通规则，先删除原规则集，再创建新规则
            await deleteRuleSet(selectedRule.id);
            await createRule(ruleData);
          } else {
            // 如果原来就是普通规则，直接更新
            await updateRule(selectedRule.id, ruleData);
          }
        } else {
          await createRule(ruleData);
        }
      }
      await fetchData();
      setSuccess(true);
      handleCloseDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      setLoading(true);
      // 在规则集列表中查找
      const ruleSet = ruleSets.find(rs => rs.id === id);
      if (ruleSet) {
        await deleteRuleSet(id);
      } else {
        await deleteRule(id);
      }
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除失败');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateRuleSet = async (id) => {
    try {
      setLoading(true);
      await updateRuleSetRules(id);
      setSuccess(true);
      await fetchData();
    } catch (err) {
      setError(err.response?.data?.error || '更新规则���失败');
    } finally {
      setLoading(false);
    }
  };

  const cardVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { 
      opacity: 1, 
      y: 0,
      transition: {
        duration: 0.3,
      }
    }
  };

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1
      }
    }
  };

  // 统一规则卡片渲染函数
  const renderRuleCard = (item, type = 'rule') => {
    const isRuleSet = type === 'ruleset';
    // 查找对应的生成规则
    const generatedRule = isRuleSet ? rules.find(rule => rule.id === `ruleset-${item.id}`) : null;

    // 格式化时间
    const formatTime = (timestamp) => {
      if (!timestamp) return '未更新';
      try {
        const date = new Date(timestamp);
        if (isNaN(date.getTime())) return '时间格式错误';
        return date.toLocaleString();
      } catch (err) {
        return '时间格式错误';
      }
    };

    const chipLabel = isRuleSet ? item.format : item.type;
    const chipBackground = () => {
      if (isRuleSet) {
        return item.format === 'domain'
          ? 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)'
          : 'linear-gradient(45deg, #42a5f5 30%, #64b5f6 90%)';
      } else {
        return item.type === 'domain'
          ? 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)'
          : item.type === 'ip'
            ? 'linear-gradient(45deg, #42a5f5 30%, #64b5f6 90%)'
            : 'linear-gradient(45deg, #66bb6a 30%, #81c784 90%)';
      }
    };

    // 获取规则来源标签
    const getSourceLabel = () => {
      if (!isRuleSet) {
        return item.type === 'remote' ? '远程规则' : '自定义规则';
      }
      return item.type === 'list' ? '列表规则集' : 'JSON规则集';
    };

    // 获取规则来源标签颜色
    const getSourceChipColor = () => {
      if (!isRuleSet) {
        return item.type === 'remote'
          ? 'linear-gradient(45deg, #ff9800 30%, #ffa726 90%)'  // 橙色渐变
          : 'linear-gradient(45deg, #9c27b0 30%, #ba68c8 90%)'; // 紫色渐变
      }
      return item.type === 'srs'
        ? 'linear-gradient(45deg, #00bcd4 30%, #4dd0e1 90%)'  // 青色渐变
        : 'linear-gradient(45deg, #ff5722 30%, #ff7043 90%)'; // 深橙色渐变
    };

    return (
      <Grid item xs={12} sm={6} md={4} lg={3} key={item.id}>
        <motion.div
          variants={cardVariants}
          initial="hidden"
          animate="visible"
        >
          <Card sx={{ height: '100%', position: 'relative', fontSize: '0.875rem' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
                <Box sx={{ flex: 1, mr: 1 }}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 500, mb: 0.25, fontSize: '0.9rem' }}>
                    {item.name}
                  </Typography>
                  {item.description && (
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.75rem' }}>
                      {item.description}
                    </Typography>
                  )}
                </Box>
                <Box sx={{ display: 'flex', flexShrink: 0 }}>
                  <IconButton
                    size="small"
                    onClick={() => handleOpenDialog(item)}
                    sx={{ p: 0.5, mr: 0.25 }}
                  >
                    <EditIcon sx={{ fontSize: '1rem' }} />
                  </IconButton>
                  {isRuleSet && (
                    <IconButton
                      size="small"
                      onClick={() => handleUpdateRuleSet(item.id)}
                      sx={{ p: 0.5, mr: 0.25 }}
                    >
                      <RefreshIcon sx={{ fontSize: '1rem' }} />
                    </IconButton>
                  )}
                  <IconButton
                    size="small"
                    onClick={() => handleDelete(item.id)}
                    sx={{ p: 0.5 }}
                  >
                    <DeleteIcon sx={{ fontSize: '1rem' }} />
                  </IconButton>
                </Box>
              </Box>

              <Stack spacing={0.75}>
                {isRuleSet && (
                  <Box>
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem', mb: 0.25 }}>
                      规则集URL
                    </Typography>
                    <Typography variant="body2" sx={{ 
                      wordBreak: 'break-all', 
                      fontSize: '0.75rem',
                      maxHeight: '2.4em',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                    }}>
                      {item.url}
                    </Typography>
                  </Box>
                )}

                {isRuleSet && generatedRule && (
                  <Box>
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem', mb: 0.25 }}>
                      规则数量
                    </Typography>
                    <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                      {generatedRule.domains?.length || generatedRule.ips?.length || 0} 条规则
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 0.25, fontSize: '0.7rem' }}>
                      规则集更新: {formatTime(item.updated_at)}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
                      规则更新: {formatTime(generatedRule.updated_at)}
                    </Typography>
                  </Box>
                )}

                <Box>
                  <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem', mb: 0.25 }}>
                    {isRuleSet ? '规则集类型' : '匹配规则'}
                  </Typography>
                  <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                    {isRuleSet 
                      ? (item.type === 'srs' ? 'Sing-box 规则集' : 'JSON 规则集')
                      : item.pattern}
                  </Typography>
                </Box>

                <Box>
                  <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem', mb: 0.25 }}>
                    {isRuleSet ? '目标出口' : '目标节点'}
                  </Typography>
                  <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                    {isRuleSet ? item.outbound : item.target}
                  </Typography>
                </Box>

                <Stack direction="row" spacing={0.5} sx={{ mt: 0.5 }}>
                  <Chip 
                    label={chipLabel}
                    size="small"
                    sx={{
                      background: chipBackground,
                      color: '#fff',
                      fontWeight: 500,
                      height: '20px',
                      '& .MuiChip-label': {
                        fontSize: '0.7rem',
                        px: 1,
                      }
                    }}
                  />
                  <Chip 
                    label={getSourceLabel()}
                    size="small"
                    sx={{
                      background: getSourceChipColor(),
                      color: '#fff',
                      fontWeight: 500,
                      height: '20px',
                      '& .MuiChip-label': {
                        fontSize: '0.7rem',
                        px: 1,
                      }
                    }}
                  />
                  <Chip 
                    label={item.enabled ? "已启用" : "已禁用"} 
                    size="small"
                    sx={item.enabled ? {
                      ...styles.chip.success,
                      background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                      color: '#fff',
                      fontWeight: 500,
                      height: '20px',
                      '& .MuiChip-label': {
                        fontSize: '0.7rem',
                        px: 1,
                      }
                    } : {
                      ...styles.chip.error,
                      background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                      color: '#fff',
                      fontWeight: 500,
                      height: '20px',
                      '& .MuiChip-label': {
                        fontSize: '0.7rem',
                        px: 1,
                      }
                    }}
                  />
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </motion.div>
      </Grid>
    );
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
                规则管理
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
                onClick={onDashboardClick}
                sx={styles.iconButton}
              >
                <OpenInNewIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">规则列表</Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
          sx={styles.actionButton}
        >
          添加规则
        </Button>
      </Box>

      <Grid container spacing={1.5}>
        {/* 只渲染规则集和非规则集生成的规则 */}
        {ruleSets.map((ruleSet) => renderRuleCard(ruleSet, 'ruleset'))}
        {rules.filter(rule => !rule.id.startsWith('ruleset-')).map((rule) => renderRuleCard(rule, 'rule'))}
      </Grid>

      <Dialog 
        open={openDialog} 
        onClose={handleCloseDialog} 
        maxWidth="xs" 
        fullWidth
        PaperProps={{
          sx: {
            position: 'fixed',
            m: 2,
            maxHeight: '80vh',
            overflowY: 'auto'
          }
        }}
      >
        <DialogTitle sx={{ pb: 1 }}>
          {selectedRule ? '编辑规则' : '添加规则'}
        </DialogTitle>
        <DialogContent sx={{ py: 1 }}>
          <Box>
            <TextField
              fullWidth
              size="small"
              label="名称"
              value={editRule.name}
              onChange={(e) => setEditRule({ ...editRule, name: e.target.value })}
              margin="normal"
              required
            />
            <FormControl fullWidth margin="normal" size="small">
              <InputLabel>规则类型</InputLabel>
              <Select
                value={editRule.type}
                onChange={(e) => {
                  const newType = e.target.value;
                  setEditRule(prev => ({
                    ...prev,
                    type: newType,
                    format: newType === 'ruleset' ? 'srs' : 'domain',
                    rules: []
                  }));
                }}
                label="规则类型"
                MenuProps={{
                  container: document.body,
                  style: { zIndex: 9999 }
                }}
              >
                {ruleTypes.map(type => (
                  <MenuItem key={type.value} value={type.value}>
                    {type.label}
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>{ruleTypes.find(t => t.value === editRule.type)?.description}</FormHelperText>
            </FormControl>

            {editRule.type === 'ruleset' && (
              <>
                <FormControl component="fieldset" margin="normal" fullWidth>
                  <FormLabel component="legend" sx={{ fontSize: '0.875rem', mb: 1 }}>规则集格式</FormLabel>
                  <RadioGroup
                    row
                    value={editRule.format}
                    onChange={(e) => setEditRule({ ...editRule, format: e.target.value })}
                  >
                    <MuiFormControlLabel 
                      value="list" 
                      control={<Radio size="small" />} 
                      label="列表格式 (.list/.txt)" 
                    />
                    <MuiFormControlLabel 
                      value="json" 
                      control={<Radio size="small" />} 
                      label="JSON格式 (.json)" 
                    />
                  </RadioGroup>
                </FormControl>
                <FormControl fullWidth margin="normal" size="small">
                  <InputLabel>规则格式</InputLabel>
                  <Select
                    value={editRule.ruleFormat}
                    onChange={(e) => setEditRule({ ...editRule, ruleFormat: e.target.value })}
                    label="规则格式"
                  >
                    <MenuItem value="domain">域名</MenuItem>
                    <MenuItem value="geoip">IP</MenuItem>
                  </Select>
                </FormControl>
                <TextField
                  fullWidth
                  size="small"
                  label="URL"
                  value={editRule.url}
                  onChange={(e) => setEditRule({ ...editRule, url: e.target.value })}
                  margin="normal"
                  required
                />
              </>
            )}

            {editRule.type === 'custom' && (
              <>
                <FormControl fullWidth margin="normal" size="small">
                  <InputLabel>规则格式</InputLabel>
                  <Select
                    value={editRule.format}
                    onChange={(e) => setEditRule({ ...editRule, format: e.target.value, rules: [] })}
                    label="规则格式"
                  >
                    <MenuItem value="domain">域名</MenuItem>
                    <MenuItem value="geoip">GeoIP</MenuItem>
                  </Select>
                </FormControl>
                <TextField
                  fullWidth
                  size="small"
                  label="规则"
                  value={editRule.rules.join('\n')}
                  onChange={(e) => setEditRule({ ...editRule, rules: e.target.value.split('\n').filter(Boolean) })}
                  margin="normal"
                  multiline
                  rows={4}
                  required
                  helperText={`每行一个${editRule.format === 'domain' ? '域名' : 'IP'}`}
                />
              </>
            )}

            <FormControl fullWidth margin="normal" size="small">
              <InputLabel>出站方式</InputLabel>
              <Select
                value={editRule.outbound}
                onChange={(e) => setEditRule({ ...editRule, outbound: e.target.value })}
                label="出站方式"
              >
                {outboundTypes.map(type => (
                  <MenuItem key={type.value} value={type.value}>
                    {type.label}
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>{outboundTypes.find(t => t.value === editRule.outbound)?.description}</FormHelperText>
            </FormControl>

            {editRule.outbound === 'node_group' && (
              <FormControl fullWidth margin="normal" size="small">
                <InputLabel>节点组</InputLabel>
                <Select
                  value={editRule.nodeGroup}
                  onChange={(e) => setEditRule({ ...editRule, nodeGroup: e.target.value })}
                  label="节点组"
                >
                  {nodeGroups.map(group => (
                    <MenuItem key={group.id} value={group.id}>
                      {group.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}

            {editRule.outbound === 'specific' && (
              <FormControl fullWidth margin="normal" size="small">
                <InputLabel>指定节点</InputLabel>
                <Select
                  value={editRule.specificNode}
                  onChange={(e) => setEditRule({ ...editRule, specificNode: e.target.value })}
                  label="指定节点"
                >
                  {nodes.map(node => (
                    <MenuItem key={node.id} value={node.id}>
                      {node.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}

            <TextField
              fullWidth
              size="small"
              label="描述"
              value={editRule.description}
              onChange={(e) => setEditRule({ ...editRule, description: e.target.value })}
              margin="normal"
              multiline
              rows={2}
            />

            <FormControlLabel
              control={
                <Switch
                  size="small"
                  checked={editRule.enabled}
                  onChange={(e) => setEditRule({ ...editRule, enabled: e.target.checked })}
                />
              }
              label="启用"
              sx={{ mt: 1 }}
            />
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={handleCloseDialog} size="small">取消</Button>
          <Button onClick={handleSaveRule} variant="contained" color="primary" size="small">
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Rules; 