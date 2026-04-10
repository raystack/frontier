'use client';

import { useRef, useState } from 'react';
import ReactCrop, {
  type Crop,
  type PixelCrop,
  type PercentCrop,
  centerCrop,
  makeAspectCrop
} from 'react-image-crop';
import { UploadIcon } from '@radix-ui/react-icons';
import {
  Avatar,
  Button,
  Dialog,
  Flex,
  IconButton,
  Text
} from '@raystack/apsara-v1';
import 'react-image-crop/dist/ReactCrop.css';
import styles from './image-upload.module.css';
import { type SyntheticEvent, type ChangeEvent } from 'react';

interface CropDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  imgSrc: string;
  onSave: (data: string) => void;
}

function CropDialog({ open, onOpenChange, imgSrc, onSave }: CropDialogProps) {
  const [crop, setCrop] = useState<Crop>();
  const imgRef = useRef<HTMLImageElement>(null);

  function onImageLoad(e: SyntheticEvent<HTMLImageElement>) {
    const { naturalWidth: width, naturalHeight: height } = e.currentTarget;
    const newCrop = centerCrop(
      makeAspectCrop(
        {
          unit: '%',
          width: 100
        },
        1,
        width,
        height
      ),
      width,
      height
    );
    setCrop(newCrop);
  }

  async function handleSave() {
    const image = imgRef.current;
    if (!image) return;

    const canvas = document.createElement('canvas');
    const scaleX = image.naturalWidth / image.width;
    const scaleY = image.naturalHeight / image.height;

    const height = ((crop?.height || 0) * image.height) / 100;
    const width = ((crop?.width || 0) * image.width) / 100;
    const x = ((crop?.x || 0) * image.width) / 100;
    const y = ((crop?.y || 0) * image.width) / 100;

    const pixelRatio = window.devicePixelRatio;
    canvas.width = width * pixelRatio;
    canvas.height = height * pixelRatio;
    const ctx = canvas.getContext('2d');

    if (!ctx) return;

    ctx.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    ctx.imageSmoothingQuality = 'high';

    ctx.drawImage(
      image,
      x * scaleX,
      y * scaleY,
      width * scaleX,
      height * scaleY,
      0,
      0,
      width,
      height
    );

    const base64Image = canvas.toDataURL('image/jpeg');
    onSave(base64Image);
    onOpenChange(false);
  }

  function handleCancel() {
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content width={600}>
        <Dialog.Header>
          <Dialog.Title>Crop your photo</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body>
          <Flex
            direction="column"
            justify="center"
            align="center"
          >
            {imgSrc ? (
              <ReactCrop
                crop={crop}
                onChange={(_: PixelCrop, percentCrop: PercentCrop) => setCrop(percentCrop)}
                aspect={1}
                className={styles.reactCrop}
                data-test-id="frontier-sdk-image-crop-preview"
              >
                <img
                  src={imgSrc}
                  alt="preview"
                  ref={imgRef}
                  onLoad={onImageLoad}
                  className={styles.previewImg}
                />
              </ReactCrop>
            ) : null}
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={handleCancel}
              data-test-id="frontier-sdk-image-crop-modal-cancel-btn"
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="accent"
              onClick={handleSave}
              data-test-id="frontier-sdk-image-crop-modal-save-btn"
            >
              Save
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
}

export interface ImageUploadProps {
  description?: string;
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  initials?: string;
  'data-test-id'?: string;
}

export function ImageUpload({
  description,
  value,
  onChange,
  disabled = false,
  initials,
  ...rest
}: ImageUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [imgSrc, setImgSrc] = useState('');
  const [showCropDialog, setShowCropDialog] = useState(false);

  function onUploadIconClick() {
    inputRef.current?.click();
  }

  function onFileChange(e: ChangeEvent<HTMLInputElement>) {
    const files = e.target.files || [];
    if (files.length > 0) {
      const file = files[0];
      const imageUrl = URL.createObjectURL(file);
      setImgSrc(imageUrl);
      setShowCropDialog(true);
      e.target.files = null;
    }
  }

  function handleSave(data: string) {
    onChange?.(data);
  }

  return (
    <Flex direction="column" gap={5} align="start" data-test-id={rest['data-test-id']}>
      <IconButton
        size={4}
        aria-label="Upload image"
        className={styles.iconButton}
        onClick={disabled ? undefined : onUploadIconClick}
        disabled={disabled}
        data-test-id="frontier-sdk-image-upload-icon"
      >
        {(value || initials) && (
          <Avatar src={value} fallback={initials} size={10} radius="small" className={styles.avatar} />
        )}
        <UploadIcon className={styles.uploadIcon} />
      </IconButton>

      {description ? (
        <Text size="regular" variant="secondary">
          {description}
        </Text>
      ) : null}

      <input
        type="file"
        accept="image/png, image/jpeg"
        ref={inputRef}
        className={styles.inputFileField}
        onChange={onFileChange}
        aria-hidden="true"
        data-test-id="frontier-sdk-image-upload-file-input"
      />

      <CropDialog
        open={showCropDialog}
        onOpenChange={setShowCropDialog}
        imgSrc={imgSrc}
        onSave={handleSave}
      />
    </Flex>
  );
}
