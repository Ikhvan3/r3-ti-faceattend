class FaceModelConfig {
  const FaceModelConfig._();

  static const identifier = 'facenet';
  static const version = 'shubham0204-facenet-2020-fp32';
  static const assetPath = 'assets/models/facenet.tflite';
  static const inputWidth = 160;
  static const inputHeight = 160;
  static const inputChannels = 3;
  static const embeddingDimension = 128;
  static const sampleTarget = 5;
  static const sampleInterval = Duration(milliseconds: 650);
  static const cropMarginRatio = 0.18;
  static const minFaceBoxRatio = 0.28;
  static const edgeMarginRatio = 0.04;
  static const maxHeadEulerY = 18.0;
  static const maxHeadEulerZ = 15.0;
}
