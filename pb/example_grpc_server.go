package pb

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	// 导入生成的代码（使用相对路径）
	generated_pb "pb-tool/grpc_output/pb"
)

// ExampleServer 实现 ExampleService 接口
type ExampleServer struct {
	generated_pb.UnimplementedExampleServiceServer
}

// GetExample 实现 GetExample 方法
func (s *ExampleServer) GetExample(ctx context.Context, req *generated_pb.GetExampleRequest) (*generated_pb.Example, error) {
	// 实现方法逻辑
	return &generated_pb.Example{
		Id:    req.Id,
		Name:  "Example",
		Value: 123,
	}, nil
}

// CreateExample 实现 CreateExample 方法
func (s *ExampleServer) CreateExample(ctx context.Context, req *generated_pb.Example) (*generated_pb.Example, error) {
	// 实现方法逻辑
	return req, nil
}

// UpdateExample 实现 UpdateExample 方法
func (s *ExampleServer) UpdateExample(ctx context.Context, req *generated_pb.Example) (*generated_pb.Example, error) {
	// 实现方法逻辑
	return req, nil
}

// DeleteExample 实现 DeleteExample 方法
func (s *ExampleServer) DeleteExample(ctx context.Context, req *generated_pb.GetExampleRequest) (*generated_pb.DeleteExampleResponse, error) {
	// 实现方法逻辑
	return &generated_pb.DeleteExampleResponse{
		Success: true,
		Message: "Example deleted",
	}, nil
}

// ListExamples 实现 ListExamples 方法
func (s *ExampleServer) ListExamples(ctx context.Context, req *generated_pb.ListExamplesRequest) (*generated_pb.ListExamplesResponse, error) {
	// 实现方法逻辑
	return &generated_pb.ListExamplesResponse{
		Examples: []*generated_pb.Example{
			{
				Id:    "1",
				Name:  "Example 1",
				Value: 123,
			},
		},
		Total:    1,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CustomServerOption 自定义服务器选项，用于配置是否只注册publish=true的方法
type CustomServerOption func(*ServerOptions)

// WithPublishOnly 只注册publish=true的方法
func WithPublishOnly(publishOnly bool) CustomServerOption {
	return func(opts *ServerOptions) {
		opts.PublishOnly = publishOnly
	}
}

// ServerOptions 服务器选项
type ServerOptions struct {
	PublishOnly bool
}

// RegisterExampleServiceWithOptions 带有publish选项的服务注册函数
func RegisterExampleServiceWithOptions(
	s *grpc.Server,
	srv generated_pb.ExampleServiceServer,
	opts ...CustomServerOption) {

	// 解析选项
	serverOpts := &ServerOptions{}
	for _, opt := range opts {
		opt(serverOpts)
	}

	// 检查是否需要根据publish选项过滤方法
	if serverOpts.PublishOnly {
		fmt.Println("注册服务：只注册 publish=true 的方法")

		// 1. 获取服务描述符
		serviceDesc := &generated_pb.ExampleService_ServiceDesc

		// 2. 遍历所有方法
		for i := range serviceDesc.Methods {
			method := &serviceDesc.Methods[i]

			// 3. 获取方法全限定名
			fullMethodName := fmt.Sprintf("/%s/%s", serviceDesc.ServiceName, method.MethodName)

			// 4. 获取方法描述符（这里需要根据实际生成的代码调整）
			// 实际使用中，您需要从生成的描述符中获取方法的选项
			// 这里为了演示，我们直接使用方法名来模拟
			fmt.Printf("检查方法: %s\n", fullMethodName)

			// 5. 根据方法名判断publish选项（从描述符动态获取）
			publish := GetPublishOptionFromMethodName(method.MethodName)

			if publish {
				fmt.Printf("✓ 方法 %s 的 publish=true，允许注册\n", fullMethodName)
				// 方法会被自动注册
			} else {
				fmt.Printf("✗ 方法 %s 的 publish=false，跳过注册\n", fullMethodName)
				// 这里可以根据需要跳过注册或做其他处理
			}
		}
	}

	// 6. 注册服务
	generated_pb.RegisterExampleServiceServer(s, srv)
	fmt.Println("服务注册完成！")
}

// GetPublishOptionFromMethodName 从文件描述符中动态获取方法的publish选项值
func GetPublishOptionFromMethodName(methodName string) bool {
	// 获取文件描述符
	fileDesc := generated_pb.File_example_proto
	if fileDesc == nil {
		fmt.Printf("警告: 未找到文件描述符，方法 %s 使用默认值 true\n", methodName)
		return true
	}

	// 遍历所有服务
	for j := 0; j < fileDesc.Services().Len(); j++ {
		service := fileDesc.Services().Get(j)
		// 将protoreflect.Name转换为string类型进行比较
		if string(service.Name()) != "ExampleService" {
			continue
		}

		// 遍历服务的所有方法
		for k := 0; k < service.Methods().Len(); k++ {
			protoMethod := service.Methods().Get(k)
			// 将protoreflect.Name转换为string类型进行比较
			if string(protoMethod.Name()) != methodName {
				continue
			}

			// 获取方法的选项
			methodOpts := protoMethod.Options()
			if methodOpts == nil {
				fmt.Printf("警告: 方法 %s 未找到选项，使用默认值 true\n", methodName)
				return true
			}

			// 使用proto.GetExtension获取publish选项值
			// 注意：proto.GetExtension函数只返回一个值
			publishValue := proto.GetExtension(methodOpts.(proto.Message), generated_pb.E_Publish)
			if publishValue == nil {
				fmt.Printf("警告: 方法 %s 未找到publish选项，使用默认值 true\n", methodName)
				return true
			}

			// 返回publish选项值
			return publishValue.(bool)
		}
	}

	// 如果未找到方法，返回默认值true
	fmt.Printf("警告: 未找到方法 %s，使用默认值 true\n", methodName)
	return true
}

// PublishInterceptor 基于publish选项的拦截器
func PublishInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 获取方法名
		fmt.Printf("拦截请求: %s\n", info.FullMethod)

		// 检查方法的publish选项
		publish := GetPublishOptionFromMethodName(getMethodNameFromFullMethod(info.FullMethod))

		// 检查是否是内部服务请求（通过metadata判断）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		// 检查是否是内部服务请求
		isInternal := md.Get("x-internal-service") != nil

		if isInternal && !publish {
			// 内部服务请求，但方法publish=false，拒绝访问
			fmt.Printf("拒绝内部服务请求: %s (publish=false)\n", info.FullMethod)
			return nil, status.Error(codes.PermissionDenied, "该接口不允许内部服务请求")
		}

		// 允许访问，继续处理
		return handler(ctx, req)
	}
}

// 从完整方法名中提取方法名
func getMethodNameFromFullMethod(fullMethod string) string {
	// 方法名格式: /service/method
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[i+1:]
		}
	}
	return fullMethod
}

func main() {
	// 创建gRPC服务器，添加publish拦截器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(PublishInterceptor()),
	)

	// 创建服务实例
	exampleServer := &ExampleServer{}

	// 注册服务，只注册publish=true的方法
	RegisterExampleServiceWithOptions(s, exampleServer, WithPublishOnly(true))

	// 启动服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("gRPC服务器启动在 :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
