import QtQuick 2.0
import Sigourney 1.0

Rectangle {
	id: canvas
	width: 640
	height: 480
	color: "black"

	Column {
		objectName: "kindColumn"
		spacing: 10
	}

	DropArea {
		anchors.fill: parent
		onDropped: {
			if (drop.source.objectName == "kind")
				ctrl.onDropKind(drop.source)
		}
	}

	property var kindComponent: Component {
		Rectangle {
			objectName: "kind"

			property string kind

			width: text.width + 10
			height: text.height + 10
			color: "red"

			Text {
				id: text
				text: kind
				x: 5
				y: 5
				color: "black"
			}

			Drag.active: mouseArea.drag.active
			property int origX
			property int origY

			MouseArea {
				id: mouseArea
				anchors.fill: parent
				drag.target: parent
				onPressed: {
					parent.origX = parent.x
					parent.origY = parent.y
				}
				onReleased: {
					parent.Drag.drop()
					parent.x = parent.origX
					parent.y = parent.origY
				}
			}
		}
	}
}
