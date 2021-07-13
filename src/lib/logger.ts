const winston = require('winston');
const config = require('../config/config');

const isTestEnvironment = config.get('/env') === 'test';

const { format, createLogger, transports } = winston;

const baseFormat = format.combine(format.timestamp(), format.padLevels());

const logFormat = format.printf(
  (info: any) => `${info.timestamp} [${info.level}] ${info.message}`
);

const logCliFormat = format.combine(baseFormat, format.colorize(), logFormat);

const logger = createLogger({
  level: 'info',
  format: logCliFormat,
  transports: [new transports.Console({ silent: isTestEnvironment })]
});

export default logger;
