import 'package:bloc/bloc.dart';
import 'package:portwhine/models/node_model.dart';

class SelectedNodeCubit extends Cubit<NodeModel?> {
  SelectedNodeCubit() : super(null);

  void setNode(NodeModel model) {
    emit(model);
  }

  void removeNode() {
    emit(null);
  }
}
