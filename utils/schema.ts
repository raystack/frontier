import Joi from 'joi';

const HTTPErrorResponse = Joi.object().keys({
  error: Joi.string().required(),
  message: Joi.string().required()
});

export const UnauthorizedResponse = HTTPErrorResponse.append({
  statusCode: Joi.number().integer().valid(401).only().required()
}).label('UnauthorizedResponse');

export const BadRequestResponse = HTTPErrorResponse.append({
  statusCode: Joi.number().integer().valid(400).only().required()
}).label('BadRequestResponse');

export const NotFoundResponse = HTTPErrorResponse.append({
  statusCode: Joi.number().integer().valid(404).only().required()
}).label('NotFoundResponse');

export const InternalServerErrorResponse = HTTPErrorResponse.append({
  statusCode: Joi.number().integer().valid(500).only().required()
}).label('InternalServerErrorResponse');
