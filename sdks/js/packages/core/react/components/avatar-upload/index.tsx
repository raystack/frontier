import ReactCrop, {
  type PixelCrop,
  type Crop,
  centerCrop,
  makeAspectCrop
} from 'react-image-crop';
import { UploadIcon } from '@radix-ui/react-icons';
import { useRef, useState } from 'react';
import { Dialog, Flex, Text, Image, Button, Avatar } from '@raystack/apsara';

import cross from '~/react/assets/cross.svg';
import 'react-image-crop/dist/ReactCrop.css';
import styles from './avatar-upload.module.css';

interface CropModalProps {
  imgSrc?: string;
  onClose: () => void;
  onSave: (data: string) => void;
}

function CropModal({ onClose, imgSrc, onSave }: CropModalProps) {
  const [crop, setCrop] = useState<Crop>();

  const imgRef = useRef<HTMLImageElement>(null);

  async function handleSave() {
    const image = imgRef.current;
    if (!image) {
      throw new Error('No Image Selected');
    }

    const canvas = document.createElement('canvas');
    const scaleX = image.naturalWidth / image.width;
    const scaleY = image.naturalHeight / image.height;

    const { height = 0, width = 0, x = 0, y = 0 } = crop || {};

    canvas.width = width;
    canvas.height = height;

    const pixelRatio = window.devicePixelRatio;
    canvas.width = width * pixelRatio;
    canvas.height = height * pixelRatio;
    const ctx = canvas.getContext('2d');

    if (!ctx) {
      throw new Error('No 2d context');
    }

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
    onClose();
  }

  function onImageLoad(e: React.SyntheticEvent<HTMLImageElement>) {
    const { naturalWidth: width, naturalHeight: height } = e.currentTarget;
    const crop = centerCrop(
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
    setCrop(crop);
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.cropModal}
      >
        <Flex
          justify="between"
          style={{
            padding: '16px 24px',
            borderBottom: '1px solid var(--border-subtle)'
          }}
        >
          <Text size={6} style={{ fontWeight: '500' }}>
            Crop your photo
          </Text>
          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={onClose}
          />
        </Flex>
        <Flex
          direction="column"
          style={{ padding: '16px 32px', maxHeight: '280px', height: '100%' }}
          justify={'center'}
          align={'center'}
        >
          {imgSrc ? (
            <ReactCrop
              crop={crop}
              onChange={c => setCrop(c)}
              aspect={1}
              className={styles.reactCrop}
            >
              <img
                src={imgSrc}
                alt="preview-pic"
                ref={imgRef}
                onLoad={onImageLoad}
                className={styles.previewImg}
              />
            </ReactCrop>
          ) : null}
        </Flex>
        <Flex
          justify="end"
          style={{
            padding: 'var(--pd-16)',
            borderTop: '1px solid var(--border-subtle)'
          }}
          gap="medium"
        >
          <Button size="medium" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button size="medium" variant="primary" onClick={handleSave}>
            Save
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}

interface AvatarUploadProps {
  subText?: string;
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  initials?: string;
}

export function AvatarUpload({
  subText,
  value,
  onChange = () => {},
  initials = '',
  disabled = false
}: AvatarUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [imgSrc, setImgSrc] = useState('');
  const [showCropModal, setShowCropModal] = useState(false);

  function onUploadIconClick() {
    const inputField = inputRef.current;
    inputField?.click();
  }

  function onFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files || [];
    if (files.length > 0) {
      const file = files[0];
      const imageUrl = URL.createObjectURL(file);
      setImgSrc(imageUrl);
      setShowCropModal(true);
      e.target.files = null;
    }
  }

  function onCloseClick() {
    setShowCropModal(false);
  }

  // disabled && value => show logo without onClick event
  // disabled && !value => show avatar with fallback
  // !disabled && value => allow user to click logo and update
  // !disabled && !value => show upload icon and update

  return (
    <div className={styles.container}>
      {disabled ? (
        <div>
          <Avatar
            src={value}
            fallback={initials}
            imageProps={{ width: '80px', height: '80px' }}
          />
        </div>
      ) : value ? (
        <div onClick={onUploadIconClick} style={{ cursor: 'pointer' }}>
          <Avatar src={value} imageProps={{ width: '80px', height: '80px' }} />
        </div>
      ) : (
        <div className={styles.uploadIconWrapper} onClick={onUploadIconClick}>
          <UploadIcon />
        </div>
      )}
      {subText ? (
        <Text style={{ color: 'var(--foreground-muted)' }}>{subText}</Text>
      ) : null}
      <input
        type="file"
        accept="image/png, image/jpeg"
        ref={inputRef}
        className={styles.inputFileField}
        onChange={onFileChange}
      />
      {showCropModal ? (
        <CropModal imgSrc={imgSrc} onClose={onCloseClick} onSave={onChange} />
      ) : null}
    </div>
  );
}
